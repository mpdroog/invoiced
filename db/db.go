// Package db implements a very simple
// filesystem abstraction to save all as
// toml and instruct Git to commit/push.
package db

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	git "github.com/go-git/go-git/v6"
	"github.com/mpdroog/invoiced/config"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Package-level variables for database state.
var (
	// Repo is the Git repository for version control.
	Repo *git.Repository
	// Path is the root path to the database directory.
	Path string
	// AlwaysLowercase forces disk I/O in lowercase for better OSX compatibility.
	AlwaysLowercase bool

	pathRegex *regexp.Regexp
	lock      *sync.RWMutex

	// OnCommit is called after successful commits with touched/moved paths.
	// Used by idx package to sync SQLite index.
	OnCommit func(touchedPaths []string, movedPaths []struct{ From, To string })
)

// Txn represents a database transaction.
type Txn struct {
	Write        bool                        // Don't toggle this bool but use the Update instead of View-func
	TouchedPaths []string                    // Paths modified during this transaction (for index sync)
	MovedPaths   []struct{ From, To string } // Paths moved during this transaction
}

// Fn is a function type for database transaction callbacks.
type Fn func(*Txn) error

// pathFilter prevents path traversal and other path-based attacks
func pathFilter(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	if strings.HasPrefix(path, "/") {
		return false
	}
	return pathRegex.Match([]byte(path))
}

// Init initializes the database at the given path, creating a Git repo if needed.
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
		return fmt.Errorf("path hack attempt: %s", path)
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

// OpenRaw opens a file and returns a raw reader.
func (t *Txn) OpenRaw(path string) (io.ReadCloser, error) {
	if !pathFilter(path) {
		return nil, fmt.Errorf("path hack attempt: %s", path)
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

// Open opens a TOML file and decodes it into the provided struct.
func (t *Txn) Open(path string, out interface{}) error {
	if !pathFilter(path) {
		return fmt.Errorf("path hack attempt: %s", path)
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
	if _, e := toml.NewDecoder(buf).Decode(out); e != nil {
		_ = file.Close()
		return e
	}
	return file.Close()
}

// OpenFirst tries to open the first existing file from the list of paths.
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

	return fmt.Errorf("file not in any given path %+v", paths)
}

// View executes a read-only transaction.
func View(fn Fn) error {
	lock.RLock()
	defer lock.RUnlock()

	txn := &Txn{Write: false}
	return fn(txn)
}
