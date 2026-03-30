// Package git provides API endpoints for git status and push operations.
package git

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/julienschmidt/httprouter"
	appconfig "github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/idx"
	"github.com/mpdroog/invoiced/middleware"
	"github.com/mpdroog/invoiced/writer"
)

// CommitInfo contains details about a single Git commit.
type CommitInfo struct {
	Hash     string `json:"hash"`
	FullHash string `json:"fullHash"`
	Message  string `json:"message"`
	Author   string `json:"author"`
	Date     string `json:"date"`
}

// StatusResponse contains unpushed commit information.
type StatusResponse struct {
	Ahead   int          `json:"ahead"`
	Commits []CommitInfo `json:"commits"`
	Remote  string       `json:"remote"`
}

// HistoryResponse contains paginated commit history.
type HistoryResponse struct {
	Commits []CommitInfo `json:"commits"`
	HasMore bool         `json:"hasMore"`
	Page    int          `json:"page"`
}

// PullPushResponse contains the result of a pull or push operation.
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
		log.Printf("git.getAuth ssh agent: %s", strconv.Quote(err.Error()))
		return nil
	}
	return auth
}

// Status returns unpushed commits (commits ahead of origin/master)
func Status(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	resp := StatusResponse{
		Ahead:   0,
		Commits: []CommitInfo{},
		Remote:  "",
	}

	// Get remote URL
	remotes, err := db.Repo.Remotes()
	if err != nil {
		log.Printf("git.Status remotes: %s", strconv.Quote(err.Error()))
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
		log.Printf("git.Status head: %s", strconv.Quote(err.Error()))
		http.Error(w, "Failed to get HEAD", http.StatusInternalServerError)
		return
	}

	// Get remote tracking branch reference
	remoteRef, err := db.Repo.Reference(plumbing.NewRemoteReferenceName("origin", "master"), true)
	if err != nil {
		// No remote tracking branch - all commits are unpushed
		// This happens when remote hasn't been fetched yet
		log.Printf("git.Status remote ref: %s (treating all as unpushed)", strconv.Quote(err.Error()))
	}

	// Get commit log
	logIter, err := db.Repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		log.Printf("git.Status log: %s", strconv.Quote(err.Error()))
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
		log.Printf("git.Status encode: %v", err)
	}
}

// Push pushes commits to the remote origin
func Push(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Push to origin
	err := db.Repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/master:refs/heads/master"},
		Auth:       getAuth(),
	})

	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			if err := writer.Encode(w, r, PullPushResponse{
				Success: true,
				Message: "Already up to date",
			}); err != nil {
				log.Printf("git.Push encode: %s", strconv.Quote(err.Error()))
			}
			return
		}

		log.Printf("git.Push: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Push failed: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.Push encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: "Pushed successfully",
	}); err != nil {
		log.Printf("git.Push encode: %s", strconv.Quote(err.Error()))
	}
}

// History returns paginated commit history
func History(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		if _, err := fmt.Sscanf(pageStr, "%d", &page); err != nil {
			log.Printf("git.History parse page: %s", err)
		}
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
		log.Printf("git.History head: %s", strconv.Quote(err.Error()))
		http.Error(w, "Failed to get HEAD", http.StatusInternalServerError)
		return
	}

	// Get commit log
	logIter, err := db.Repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		log.Printf("git.History log: %s", strconv.Quote(err.Error()))
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
		log.Printf("git.History encode: %s", strconv.Quote(err.Error()))
	}
}

// DiscardAll resets the repository to origin/master (discards all local changes)
func DiscardAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get remote tracking branch reference
	remoteRef, err := db.Repo.Reference(plumbing.NewRemoteReferenceName("origin", "master"), true)
	if err != nil {
		log.Printf("git.DiscardAll remote ref: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to find origin/master: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.DiscardAll encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Get worktree
	tree, err := db.Repo.Worktree()
	if err != nil {
		log.Printf("git.DiscardAll worktree: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Failed: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.DiscardAll encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Hard reset to origin/master
	err = tree.Reset(&git.ResetOptions{
		Commit: remoteRef.Hash(),
		Mode:   git.HardReset,
	})

	if err != nil {
		log.Printf("git.DiscardAll reset: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Reset failed: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.DiscardAll encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Rebuild search index after reset
	if err := idx.Rebuild(appconfig.DbPath); err != nil {
		log.Printf("git.DiscardAll rebuild index: %s", strconv.Quote(err.Error()))
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: "Discarded all local changes",
	}); err != nil {
		log.Printf("git.DiscardAll encode: %s", strconv.Quote(err.Error()))
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
		log.Printf("git.ResetTo commit not found: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Commit not found: %s", hash),
		}); err != nil {
			log.Printf("git.ResetTo encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Get worktree
	tree, err := db.Repo.Worktree()
	if err != nil {
		log.Printf("git.ResetTo worktree: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Failed: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.ResetTo encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Hard reset to the commit
	err = tree.Reset(&git.ResetOptions{
		Commit: commitHash,
		Mode:   git.HardReset,
	})

	if err != nil {
		log.Printf("git.ResetTo reset: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Reset failed: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.ResetTo encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Rebuild search index after reset
	if err := idx.Rebuild(appconfig.DbPath); err != nil {
		log.Printf("git.ResetTo rebuild index: %s", strconv.Quote(err.Error()))
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: fmt.Sprintf("Reset to commit %s", hash[:7]),
	}); err != nil {
		log.Printf("git.ResetTo encode: %s", strconv.Quote(err.Error()))
	}
}

// Pull fetches and merges changes from the remote origin
func Pull(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get worktree
	tree, err := db.Repo.Worktree()
	if err != nil {
		log.Printf("git.Pull worktree: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Pull failed: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.Pull encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Pull from origin
	err = tree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       getAuth(),
	})

	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			if err := writer.Encode(w, r, PullPushResponse{
				Success: true,
				Message: "Already up to date",
			}); err != nil {
				log.Printf("git.Pull encode: %s", strconv.Quote(err.Error()))
			}
			return
		}

		log.Printf("git.Pull: %s", strconv.Quote(err.Error()))
		if err := writer.Encode(w, r, PullPushResponse{
			Success: false,
			Message: fmt.Sprintf("Pull failed: %s", strconv.Quote(err.Error())),
		}); err != nil {
			log.Printf("git.Pull encode: %s", strconv.Quote(err.Error()))
		}
		return
	}

	// Rebuild search index after pull
	if err := idx.Rebuild(appconfig.DbPath); err != nil {
		log.Printf("git.Pull rebuild index: %s", strconv.Quote(err.Error()))
	}

	if err := writer.Encode(w, r, PullPushResponse{
		Success: true,
		Message: "Pulled successfully",
	}); err != nil {
		log.Printf("git.Pull encode: %s", strconv.Quote(err.Error()))
	}
}
