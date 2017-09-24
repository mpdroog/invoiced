// Package db implements a very simple
// filesystem abstraction to save all as
// toml and instruct Git to commit/push.
package db

import (
	"os"
	"bufio"
	git "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	"github.com/BurntSushi/toml"
	"regexp"
	"log"
	"strings"
	"fmt"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"time"
	"github.com/mpdroog/invoiced/config"
	"sync"
	"io"
)

var (
	Repo *git.Repository
	Path string
	AlwaysLowercase bool // Force disk I/O in lowercase for better OSX compatibility

	canPush chan struct{}
	pathRegex *regexp.Regexp
	lock *sync.RWMutex
)

type Txn struct {
	Write bool // Don't toggle this bool but use the Update instead of View-func
}
type Fn func(*Txn) error

// Simple path hack prevention
func pathFilter(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	return pathRegex.Match([]byte(path))
}

func Init(path string) error {
	lock = new(sync.RWMutex)
	lock.RLock()
	defer lock.RUnlock()

	Path = path
	if !strings.HasSuffix(Path, "/") {
		Path += "/"
	}
	pathRegex = regexp.MustCompile(`^[A-Za-z0-9_\-\/{}]+(.toml)?$`)
	canPush = make(chan struct{}, 10) // non-blocking

	if !pathFilter(Path) {
		return fmt.Errorf("Path hack attempt: %s", path)
	}

	if _, e := os.Stat(Path+".git"); os.IsNotExist(e) {
		if config.Verbose {
			log.Printf("Create git-repo")
		}
		repo, e := git.PlainInit(path, false)
		if e != nil {
			return e
		}
		Repo = repo

		cfg := &gitconfig.RemoteConfig{
		    Name: "github",
		    URL:  "https://github.com/mpdroog/acct.git",
		}
		if e := cfg.Validate(); e != nil {
			return e
		}
		if _, e = Repo.CreateRemote(cfg); e != nil {
			return e
		}

		// TODO: create basic file structure?

	} else {
		if config.Verbose {
			log.Printf("Load git-repo")
		}
		repo, e := git.PlainOpen(path)
		if e != nil {
			return e
		}
		Repo = repo

		// revert any unchanged files
		// (added to support application crash)
		if e := revert(); e != nil {
			return e
		}
	}

	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	opts := &git.PullOptions{RemoteName: "github", Auth: http.NewBasicAuth("mpdroog", "2dqqR24m")}
	if e := opts.Validate(); e != nil {
		return e
	}

	// Async git push..
	go func() {
		// initial pull
		if e := tree.Pull(opts); e != nil {
			// ignore empty repo warnings
			if e.Error() != "remote repository is empty" && e.Error() != "already up-to-date" {
				log.Printf("[Git] Pull err=%s", e.Error())
			}
		}

		// err = small fn to delay retry of push
		err := func() {
			time.Sleep(time.Minute * 3)
			canPush <- struct{}{}
		}
		for {
			if config.Verbose {
				log.Printf("Git awaiting push")
			}
			<- canPush
			repos, e := Repo.Remotes()
			if e != nil {
				log.Printf("[Git] Remotes fail=%s", e.Error())
				err()
				continue
			}
			for _, repo := range repos {
				if config.Verbose {
					log.Printf("Push to %s", repo.Config().Name)
				}
				opts := &git.PushOptions{RemoteName: repo.Config().Name, Auth: http.NewBasicAuth("mpdroog", "2dqqR24m")}
				if e := opts.Validate(); e != nil {
					log.Printf("[Git] Push validate=%s", e.Error())
					err()
				}
				if e := repo.Push(opts); e != nil {
					log.Printf("[Git] Push fail=%s", e.Error())
					err()
				}
			}
		}
	}()

	return nil
}

/*func AddRemote(name, url string) error {
	_, e := Repo.CreateRemote(&config.RemoteConfig{
	    Name: name,
	    URL:  url,
	})
	return e
}

func Remotes() ([]string, error) {
	var out []string

	for _, repo := range Repo.Remotes() {
		out = append(out, repo.Config().URL)
	}

	return out, nil
}*/

func (t *Txn) OpenRaw(path string) (io.ReadCloser, error) {
	if !pathFilter(path) {
		return nil, fmt.Errorf("Path hack attempt: %s", path)
	}
	if AlwaysLowercase {
		path = strings.ToLower(path)
	}

	tree, e := Repo.Worktree()
	if e != nil {
		return nil, e
	}
	return tree.Filesystem.Open(path)
}

func (t *Txn) Open(path string, out interface{}) (error) {
	if !pathFilter(path) {
		return fmt.Errorf("Path hack attempt: %s", path)
	}
	if AlwaysLowercase {
		path = strings.ToLower(path)
	}

	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	file, e := tree.Filesystem.Open(path)
	if e != nil {
		return e
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	if _, e := toml.DecodeReader(buf, out); e != nil {
		return e
	}
	return nil
}

func (t *Txn) OpenFirst(paths []string, out interface{}) (error) {
	for _, path := range paths {
		e := t.Open(path, out)
		if e != nil {
			if os.IsNotExist(e) {
				// try next file!
				continue
			}
			return e
		}
		return nil
	}

	return fmt.Errorf("File not in any given path %+v", paths)
}

func View(fn Fn) error {
	lock.Lock()
	defer lock.Unlock()

	txn := &Txn{Write: false}
	return fn(txn)
}