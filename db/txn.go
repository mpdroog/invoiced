package db

import (
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"path"
	"os"
	"fmt"
	"bufio"
	git "gopkg.in/src-d/go-git.v4"
	"github.com/BurntSushi/toml"
	"time"
)

type Commit struct {
	Name    string
	Email   string
	Message string
}

func (t *Txn) Save(file string, in interface{}) error {
	if !t.Write {
		panic("DevErr: Save-func only allowed in Update")
	}
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

func (t *Txn) Remove(path string) error {
	if !t.Write {
		panic("DevErr: Remove-func only allowed in Update")
	}
	if !pathFilter(path) {
		return fmt.Errorf("Path hack attempt: %s", path)
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