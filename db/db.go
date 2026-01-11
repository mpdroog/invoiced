// Package db implements a very simple
// filesystem abstraction to save all as
// toml and instruct Git to commit/push.
package db

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/mpdroog/invoiced/config"
	git "gopkg.in/src-d/go-git.v4"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	Repo            *git.Repository
	Path            string
	AlwaysLowercase bool // Force disk I/O in lowercase for better OSX compatibility

	pathRegex *regexp.Regexp
	lock      *sync.RWMutex
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
	pathRegex = regexp.MustCompile(`^[A-Za-z0-9\._\-\/{}]+$`)

	if !pathFilter(Path) {
		return fmt.Errorf("Path hack attempt: %s", path)
	}

	if _, e := os.Stat(Path + ".git"); os.IsNotExist(e) {
		if config.Verbose {
			log.Printf("Create git-repo")
		}
		repo, e := git.PlainInit(path, false)
		if e != nil {
			return e
		}
		Repo = repo

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

		if config.Verbose {
			log.Printf("Revert outstanding")
		}
		// revert any unchanged files
		// (added to support application crash)
		if e := revert(); e != nil {
			return e
		}
	}

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

func (t *Txn) Open(path string, out interface{}) error {
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

	buf := bufio.NewReader(file)
	if _, e := toml.DecodeReader(buf, out); e != nil {
		file.Close() /* ignore err, write err takes precedence */
		return e
	}
	return file.Close()
}

func (t *Txn) OpenFirst(paths []string, out interface{}) error {
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
	lock.RLock()
	defer lock.RUnlock()

	txn := &Txn{Write: false}
	return fn(txn)
}
