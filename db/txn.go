package db

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/src-d/go-billy.v3"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"os"
	"path"
	"strings"
	"time"
)

type Commit struct {
	Name    string
	Email   string
	Message string
}

func (t *Txn) Save(file string, isNew bool, in interface{}) error {
	if !t.Write {
		panic("DevErr: Save-func only allowed in Update")
	}
	if !pathFilter(file) {
		return fmt.Errorf("Path hack attempt: %s", file)
	}
	if AlwaysLowercase {
		file = strings.ToLower(file)
	}

	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}

	// Ensure dirs exist
	if e := tree.Filesystem.MkdirAll(path.Dir(file), os.ModePerm); e != nil {
		return e
	}

	var (
		f billy.File
	)
	if isNew {
		if _, e := tree.Filesystem.Stat(file); !os.IsNotExist(e) {
			return fmt.Errorf("File already exists %s", file)
		}
		f, e = tree.Filesystem.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	} else {
		f, e = tree.Filesystem.OpenFile(file, os.O_RDWR, 0755)
	}
	if e != nil {
		return e
	}

	buf := bufio.NewWriter(f)
	if e := toml.NewEncoder(buf).Encode(in); e != nil {
		f.Close() /* ignore err, write err takes precedence */
		return e
	}
	if e := buf.Flush(); e != nil {
		f.Close() /* ignore err, write err takes precedence */
		return e
	}

	// commit on git
	if _, e := tree.Add(file); e != nil {
		f.Close() /* ignore err, write err takes precedence */
		return e
	}
	return f.Close()
}

func (t *Txn) Remove(path string) error {
	if !t.Write {
		panic("DevErr: Remove-func only allowed in Update")
	}
	if !pathFilter(path) {
		return fmt.Errorf("Path hack attempt: %s", path)
	}
	if AlwaysLowercase {
		path = strings.ToLower(path)
	}

	// commit on git
	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	if _, e := tree.Remove(path); e != nil {
		return e
	}
	return nil
}

func (t *Txn) Move(from, to string) error {
	if !t.Write {
		panic("DevErr: Move-func only allowed in Update")
	}
	if !pathFilter(from) {
		return fmt.Errorf("Path hack attempt: %s", from)
	}
	if !pathFilter(to) {
		return fmt.Errorf("Path hack attempt: %s", to)
	}
	if AlwaysLowercase {
		from = strings.ToLower(from)
		to = strings.ToLower(to)
	}

	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	if _, e := tree.Move(from, to); e != nil {
		return e
	}
	return nil
}

func commit(msg, name, email string) error {
	// TODO: Set When to something consistent?
	opts := &git.CommitOptions{Author: &object.Signature{Name: name, Email: email, When: time.Now()}}
	if e := opts.Validate(Repo); e != nil {
		return e
	}
	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	if _, e := tree.Commit(msg, opts); e != nil {
		return e
	}

	// push
	canPush <- struct{}{}
	return nil
}

func revert() error {
	tree, e := Repo.Worktree()
	if e != nil {
		return e
	}
	return tree.Reset(&git.ResetOptions{Mode: git.HardReset})
}

func Update(change Commit, fn Fn) error {
	lock.RLock()
	defer lock.RUnlock()

	txn := &Txn{Write: true}
	if e := fn(txn); e != nil {
		if se := revert(); se != nil {
			e = fmt.Errorf("%s (%s)", e.Error(), se.Error())
		}
		return e
	}
	return commit(change.Message, change.Name, change.Email)
}
