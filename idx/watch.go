// Package idx watch.go provides background monitoring for external git changes.
// When HEAD changes (e.g., from a manual git pull), it triggers an index rebuild.
package idx

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
)

var (
	lastHead     string
	lastHeadLock sync.RWMutex
	stopChan     chan struct{}
)

// getHead returns the current git HEAD commit hash.
func getHead() (string, error) {
	ref, err := db.Repo.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}

// getStoredHead retrieves the HEAD hash stored in SQLite from the last index build.
func getStoredHead() (string, error) {
	var value string
	err := DB.QueryRowContext(context.Background(),
		"SELECT value FROM metadata WHERE key = 'indexed_head'").Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil // No stored HEAD yet
	}
	return value, err
}

// saveHead stores the current HEAD hash in SQLite.
func saveHead(head string) error {
	_, err := DB.ExecContext(context.Background(),
		"INSERT OR REPLACE INTO metadata (key, value) VALUES ('indexed_head', ?)", head)
	return err
}

// updateLastHead stores the current HEAD hash in memory and SQLite.
func updateLastHead() error {
	head, err := getHead()
	if err != nil {
		return err
	}
	lastHeadLock.Lock()
	lastHead = head
	lastHeadLock.Unlock()

	// Also persist to SQLite for cross-restart detection
	return saveHead(head)
}

// StartWatcher starts a background goroutine that periodically checks if git HEAD
// has changed (e.g., from a manual git pull) and rebuilds the index if needed.
// It also checks immediately on startup for changes that occurred while server was stopped.
func StartWatcher(dbPath string) error {
	// Check if index is stale from before server restart
	currentHead, err := getHead()
	if err != nil {
		return err
	}

	storedHead, err := getStoredHead()
	if err != nil {
		return err
	}

	// If stored HEAD differs from current, rebuild immediately
	if storedHead != "" && storedHead != currentHead {
		log.Printf("idx: HEAD changed while server was stopped (stored: %.8s, current: %.8s), rebuilding index...", storedHead, currentHead)
		if err := Rebuild(dbPath); err != nil {
			return err
		}
	}

	// Store current HEAD in memory
	lastHeadLock.Lock()
	lastHead = currentHead
	lastHeadLock.Unlock()

	// Ensure HEAD is saved (in case this is first run or rebuild didn't save it)
	if err := saveHead(currentHead); err != nil {
		log.Printf("idx: warning: failed to save HEAD: %v", err)
	}

	stopChan = make(chan struct{})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				checkAndRebuild(dbPath)
			}
		}
	}()

	if config.Verbose {
		log.Printf("idx: HEAD watcher started (checking every 5s)")
	}
	return nil
}

// StopWatcher stops the background watcher goroutine.
func StopWatcher() {
	if stopChan != nil {
		close(stopChan)
	}
}

// checkAndRebuild compares current HEAD with stored HEAD and rebuilds if changed.
func checkAndRebuild(dbPath string) {
	currentHead, err := getHead()
	if err != nil {
		log.Printf("idx.checkAndRebuild: failed to get HEAD: %s", err)
		return
	}

	lastHeadLock.RLock()
	changed := currentHead != lastHead
	lastHeadLock.RUnlock()

	if changed {
		log.Printf("idx: HEAD changed (external git operation detected), rebuilding index...")
		if err := Rebuild(dbPath); err != nil {
			log.Printf("idx.checkAndRebuild: rebuild failed: %s", err)
			return
		}

		// Update stored HEAD after successful rebuild
		lastHeadLock.Lock()
		lastHead = currentHead
		lastHeadLock.Unlock()

		log.Printf("idx: index rebuilt successfully after external git change")
	}
}
