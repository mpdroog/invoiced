package db

import (
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type CommitMessage struct {
	*object.Commit
}

func (t *Txn) Logs(limit int, fn func(c *CommitMessage) error) (error) {
    ref, e := Repo.Head()
    if e != nil {
    	return e
    }

    cIter, e := Repo.Log(&git.LogOptions{From: ref.Hash()})
    if e != nil {
    	return e
    }

    i := 0
    e = cIter.ForEach(func(c *object.Commit) error {
    	i++
    	if limit+1 == i {
    		return storer.ErrStop
    	}
        return fn(&CommitMessage{c})
    })
    return e
}