// Package db implements a very simple
// filesystem abstraction to save all as
// toml and instruct Git to commit/push.
package db

import (
	"os"
	"bufio"
	git "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	//"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"github.com/BurntSushi/toml"
	//"path/filepath"
	"path"
	"regexp"
	"log"
	"strings"
	"fmt"
	"io/ioutil"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"time"
	"github.com/mpdroog/invoiced/config"
)

var (
	Repo *git.Repository
	Path string
	canPush chan struct{}
	pathRegex *regexp.Regexp
)

// Simple path hack prevention
func pathFilter(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	return pathRegex.Match([]byte(path))
}

func Init(path string) error {
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

	} else {
		if config.Verbose {
			log.Printf("Load git-repo")
		}
		repo, e := git.PlainOpen(path)
		if e != nil {
			return e
		}
		Repo = repo
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

func Open(path string, out interface{}) (error) {
	if !pathFilter(path) {
		return fmt.Errorf("Path hack attempt: %s", path)
	}

	abs := Path+path
	file, e := os.Open(abs)
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

func OpenFirst(paths []string, out interface{}) (error) {
	for _, path := range paths {
		if !pathFilter(path) {
			return fmt.Errorf("Path hack attempt: %s", path)
		}

		abs := Path+path
		file, e := os.Open(abs)
		if e != nil {
			if os.IsNotExist(e) {
				// try next file!
				continue
			}
			return e
		}
		defer file.Close()

		buf := bufio.NewReader(file)
		if _, e := toml.DecodeReader(buf, out); e != nil {
			return e
		}
		return nil
	}

	return fmt.Errorf("No file found")
}

func Save(file string, in interface{}) error {
	if !pathFilter(file) {
		return fmt.Errorf("Path hack attempt: %s", file)
	}

	abs := Path+file
	// ensure dir exists
	if e := os.MkdirAll(path.Dir(abs), os.ModePerm); e != nil {
		return e
	}

	// TODO: Flag to determine if overwrite is allowed? isNew?

	// overwrite file
	f, e := os.OpenFile(abs, os.O_RDWR|os.O_CREATE, 0755)
	if e != nil {
		return e
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	if e := toml.NewEncoder(buf).Encode(in); e != nil {
		return e
	}
	if e := buf.Flush(); e != nil {
		return e
	}

	// commit on git
	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	if _, e := tree.Add(file); e != nil {
		return e
	}
	return nil
}

func Remove(path string) error {
	if !pathFilter(path) {
		return fmt.Errorf("Path hack attempt: %s", path)
	}
	abs := Path+path
	if e := os.Remove(abs); e != nil {
		return e
	}

	// commit on git
	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	if _, e := tree.Add(path); e != nil {
		return e
	}
	return nil
}

// TODO: Use this..
func Commit() error {
	// TODO: Set When to something consistent?
	opts := &git.CommitOptions{Author: &object.Signature{Name: "MP Droog", Email: "rootdev@gmail.com", When: time.Now()}}
	if e := opts.Validate(Repo); e != nil {
		return e
	}
	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	if _, e := tree.Commit("Autocommit", opts); e != nil {
		return e
	}

	// push
	canPush <- struct{}{}
	return nil
}

// Find file
// TODO: Remove this hack..
func Lookup(path, fileName string) (string, error) {
	if !pathFilter(path) {
		return "", fmt.Errorf("Path hack attempt: %s", path)
	}
	abs := Path+path

	files, e := ioutil.ReadDir(abs)
	if e != nil {
		return "", e
	}

	out := ""
	for _, file := range files {
		p := fmt.Sprintf("%s/%s/%s.toml", abs, file, fileName)
		if _, e := os.Stat(p); e == nil {
			if out != "" {
				return "", fmt.Errorf("Duplicate filename=" + fileName)
			}
			out = p
		}
	}
	return out, nil
}