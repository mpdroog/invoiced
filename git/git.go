// Package git provides API endpoints for git status and push operations.
package git

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/middleware"
	"github.com/mpdroog/invoiced/writer"
)

type CommitInfo struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
}

type StatusResponse struct {
	Ahead   int          `json:"ahead"`
	Commits []CommitInfo `json:"commits"`
	Remote  string       `json:"remote"`
}

type PullPushResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// getAuth returns the appropriate auth method based on remote URL and config
func getAuth() transport.AuthMethod {
	user, token := middleware.GitCredentials()

	// If credentials configured, use HTTP basic auth
	if user != "" && token != "" {
		return &githttp.BasicAuth{
			Username: user,
			Password: token,
		}
	}

	// Try SSH agent for git@ remotes
	auth, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		log.Printf("git.getAuth ssh agent: %s", err.Error())
		return nil
	}
	return auth
}

// getRemoteURL returns the first remote URL
func getRemoteURL() string {
	remotes, err := db.Repo.Remotes()
	if err != nil || len(remotes) == 0 {
		return ""
	}
	urls := remotes[0].Config().URLs
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

// isHTTPRemote checks if the remote URL is HTTP/HTTPS
func isHTTPRemote() bool {
	url := getRemoteURL()
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// Status returns unpushed commits (commits ahead of origin/master)
func Status(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	resp := StatusResponse{
		Ahead:   0,
		Commits: []CommitInfo{},
		Remote:  "",
	}

	// Get remote URL
	remotes, err := db.Repo.Remotes()
	if err != nil {
		log.Printf("git.Status remotes: %s", err.Error())
		http.Error(w, "Failed to get remotes", http.StatusInternalServerError)
		return
	}
	if len(remotes) > 0 {
		urls := remotes[0].Config().URLs
		if len(urls) > 0 {
			resp.Remote = urls[0]
		}
	}

	// Get HEAD reference
	head, err := db.Repo.Head()
	if err != nil {
		log.Printf("git.Status head: %s", err.Error())
		http.Error(w, "Failed to get HEAD", http.StatusInternalServerError)
		return
	}

	// Get remote tracking branch reference
	remoteRef, err := db.Repo.Reference(plumbing.NewRemoteReferenceName("origin", "master"), true)
	if err != nil {
		// No remote tracking branch - all commits are unpushed
		// This happens when remote hasn't been fetched yet
		log.Printf("git.Status remote ref: %s (treating all as unpushed)", err.Error())
	}

	// Get commit log
	logIter, err := db.Repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		log.Printf("git.Status log: %s", err.Error())
		http.Error(w, "Failed to get log", http.StatusInternalServerError)
		return
	}
	defer logIter.Close()

	// Iterate through commits until we hit the remote ref
	for {
		commit, err := logIter.Next()
		if err != nil {
			break
		}

		// Stop if we've reached the remote tracking branch
		if remoteRef != nil && commit.Hash == remoteRef.Hash() {
			break
		}

		resp.Commits = append(resp.Commits, CommitInfo{
			Hash:    commit.Hash.String()[:7],
			Message: commit.Message,
			Author:  commit.Author.Name,
			Date:    commit.Author.When.Format("2006-01-02 15:04"),
		})
		resp.Ahead++

		// Limit to 100 commits
		if resp.Ahead >= 100 {
			break
		}
	}

	if err := writer.Encode(w, r, resp); err != nil {
		log.Printf("git.Status encode: %s", err.Error())
	}
}

// Push pushes commits to the remote origin
func Push(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Push to origin
	err := db.Repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/master:refs/heads/master"},
		Auth:       getAuth(),
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			if err := writer.Encode(w, r, PullPushResponse{
				Success: true,
				Message: "Already up to date",
			}); err != nil {
				log.Printf("git.Push encode: %s", err.Error())
			}
			return
		}

		log.Printf("git.Push: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Push failed: %s", err.Error()),
		}); err != nil {
			log.Printf("git.Push encode: %s", err.Error())
		}
		return
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: "Pushed successfully",
	}); err != nil {
		log.Printf("git.Push encode: %s", err.Error())
	}
}

// Pull fetches and merges changes from the remote origin
func Pull(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get worktree
	tree, err := db.Repo.Worktree()
	if err != nil {
		log.Printf("git.Pull worktree: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Pull failed: %s", err.Error()),
		}); err != nil {
			log.Printf("git.Pull encode: %s", err.Error())
		}
		return
	}

	// Pull from origin
	err = tree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       getAuth(),
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			if err := writer.Encode(w, r, PullPushResponse{
				Success: true,
				Message: "Already up to date",
			}); err != nil {
				log.Printf("git.Pull encode: %s", err.Error())
			}
			return
		}

		log.Printf("git.Pull: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Pull failed: %s", err.Error()),
		}); err != nil {
			log.Printf("git.Pull encode: %s", err.Error())
		}
		return
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: "Pulled successfully",
	}); err != nil {
		log.Printf("git.Pull encode: %s", err.Error())
	}
}
