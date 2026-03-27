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
	Hash     string `json:"hash"`
	FullHash string `json:"fullHash"`
	Message  string `json:"message"`
	Author   string `json:"author"`
	Date     string `json:"date"`
}

type StatusResponse struct {
	Ahead   int          `json:"ahead"`
	Commits []CommitInfo `json:"commits"`
	Remote  string       `json:"remote"`
}

type HistoryResponse struct {
	Commits []CommitInfo `json:"commits"`
	HasMore bool         `json:"hasMore"`
	Page    int          `json:"page"`
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
			Hash:     commit.Hash.String()[:7],
			FullHash: commit.Hash.String(),
			Message:  commit.Message,
			Author:   commit.Author.Name,
			Date:     commit.Author.When.Format("2006-01-02 15:04"),
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

// History returns paginated commit history
func History(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	if page < 0 {
		page = 0
	}

	const perPage = 20
	skip := page * perPage

	resp := HistoryResponse{
		Commits: []CommitInfo{},
		HasMore: false,
		Page:    page,
	}

	// Get HEAD reference
	head, err := db.Repo.Head()
	if err != nil {
		log.Printf("git.History head: %s", err.Error())
		http.Error(w, "Failed to get HEAD", http.StatusInternalServerError)
		return
	}

	// Get commit log
	logIter, err := db.Repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		log.Printf("git.History log: %s", err.Error())
		http.Error(w, "Failed to get log", http.StatusInternalServerError)
		return
	}
	defer logIter.Close()

	// Skip to the right page
	skipped := 0
	for skipped < skip {
		_, err := logIter.Next()
		if err != nil {
			break
		}
		skipped++
	}

	// Collect commits for this page
	count := 0
	for count < perPage+1 { // +1 to check if there are more
		commit, err := logIter.Next()
		if err != nil {
			break
		}

		if count < perPage {
			resp.Commits = append(resp.Commits, CommitInfo{
				Hash:     commit.Hash.String()[:7],
				FullHash: commit.Hash.String(),
				Message:  strings.TrimSpace(commit.Message),
				Author:   commit.Author.Name,
				Date:     commit.Author.When.Format("2006-01-02 15:04"),
			})
		} else {
			resp.HasMore = true
		}
		count++
	}

	if err := writer.Encode(w, r, resp); err != nil {
		log.Printf("git.History encode: %s", err.Error())
	}
}

// DiscardAll resets the repository to origin/master (discards all local changes)
func DiscardAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get remote tracking branch reference
	remoteRef, err := db.Repo.Reference(plumbing.NewRemoteReferenceName("origin", "master"), true)
	if err != nil {
		log.Printf("git.DiscardAll remote ref: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to find origin/master: %s", err.Error()),
		}); err != nil {
			log.Printf("git.DiscardAll encode: %s", err.Error())
		}
		return
	}

	// Get worktree
	tree, err := db.Repo.Worktree()
	if err != nil {
		log.Printf("git.DiscardAll worktree: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Failed: %s", err.Error()),
		}); err != nil {
			log.Printf("git.DiscardAll encode: %s", err.Error())
		}
		return
	}

	// Hard reset to origin/master
	err = tree.Reset(&git.ResetOptions{
		Commit: remoteRef.Hash(),
		Mode:   git.HardReset,
	})

	if err != nil {
		log.Printf("git.DiscardAll reset: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Reset failed: %s", err.Error()),
		}); err != nil {
			log.Printf("git.DiscardAll encode: %s", err.Error())
		}
		return
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: "Discarded all local changes",
	}); err != nil {
		log.Printf("git.DiscardAll encode: %s", err.Error())
	}
}

// ResetTo resets the repository to a specific commit (hard reset)
func ResetTo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	hash := ps.ByName("hash")
	if hash == "" {
		http.Error(w, "Missing commit hash", http.StatusBadRequest)
		return
	}

	// Resolve the hash
	commitHash := plumbing.NewHash(hash)

	// Verify commit exists
	_, err := db.Repo.CommitObject(commitHash)
	if err != nil {
		log.Printf("git.ResetTo commit not found: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Commit not found: %s", hash),
		}); err != nil {
			log.Printf("git.ResetTo encode: %s", err.Error())
		}
		return
	}

	// Get worktree
	tree, err := db.Repo.Worktree()
	if err != nil {
		log.Printf("git.ResetTo worktree: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Failed: %s", err.Error()),
		}); err != nil {
			log.Printf("git.ResetTo encode: %s", err.Error())
		}
		return
	}

	// Hard reset to the commit
	err = tree.Reset(&git.ResetOptions{
		Commit: commitHash,
		Mode:   git.HardReset,
	})

	if err != nil {
		log.Printf("git.ResetTo reset: %s", err.Error())
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Reset failed: %s", err.Error()),
		}); err != nil {
			log.Printf("git.ResetTo encode: %s", err.Error())
		}
		return
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: fmt.Sprintf("Reset to commit %s", hash[:7]),
	}); err != nil {
		log.Printf("git.ResetTo encode: %s", err.Error())
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
