// Package config handles application configuration loading and global settings.
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Global configuration variables.
var (
	Verbose    bool   // Verbose enables debug logging
	DbPath     string // DbPath is the path to the database directory
	HTTPListen string // HTTPListen is the HTTP server listen address
	CurDir     string // CurDir is the current working directory

	Hostname string // Hostname is the system hostname
	C        Config // C is the loaded configuration
)

// Queue defines SMTP queue settings for sending emails.
type Queue struct {
	User      string
	Pass      string
	Host      string
	Port      int
	From      string
	FromReply string
	Display   string
	Subject   string
	BCC       []string
}

// Config is the main configuration structure.
type Config struct {
	Queues map[string]Queue
}

// Ignore returns true if the given filename should be ignored.
func Ignore(str string) bool {
	return strings.ToLower(str) == ".ds_store"
}

// Open loads the configuration from the given TOML file.
func Open(f string) error {
	r, e := os.Open(filepath.Clean(f))
	if e != nil {
		return e
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("config.Open close: %s", err)
		}
	}()
	if _, e := toml.NewDecoder(r).Decode(&C); e != nil {
		return fmt.Errorf("TOML: %w", e)
	}

	if Verbose {
		log.Printf("C=%+v\n", C)
	}

	Hostname, e = os.Hostname()
	if e != nil {
		return e
	}
	return nil
}
