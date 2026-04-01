package db

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/mpdroog/invoiced/config"
)

const maxFiles = 1000
const deadline = 5 * time.Second

// Pagination specifies offset and limit for list queries.
type Pagination struct {
	From  int // TODO: possible?
	Count int
}

// PaginationHeader contains the total count for paginated results.
type PaginationHeader struct {
	Total int
}

// RawList returns directory entries for the given path.
func (t *Txn) RawList(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

// List iterates over TOML files in the given paths and calls f for each.
func (t *Txn) List(path []string, p Pagination, mem interface{}, f func(string, string, string) error) (PaginationHeader, error) {
	var page PaginationHeader

	if p.Count > maxFiles {
		return page, fmt.Errorf("pagination.Count exceeds maxFiles(%d)", maxFiles)
	}
	timeout := time.Now().Add(deadline)

	var pfPath []string
	for _, p := range path {
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		if AlwaysLowercase {
			p = strings.ToLower(p)
		}
		pfPath = append(pfPath, Path+p)
	}

	paths, e := parseWildcards(pfPath)
	if e != nil {
		return page, e
	}

	i := 0
	for _, path := range paths {
		if time.Now().After(timeout) {
			return page, fmt.Errorf("deadline timeout")
		}
		// TODO: pagination??

		files, e := os.ReadDir(path)
		if os.IsNotExist(e) {
			// No such folder (git only tracks folders when they have content)
			continue
		}
		if e != nil {
			return page, e
		}

		abs := path
		for _, file := range files {
			if time.Now().After(timeout) {
				return page, fmt.Errorf("deadline timeout")
			}

			if file.IsDir() {
				if config.Verbose {
					log.Printf("Ignore %s (is directory)\n", abs+file.Name())
				}
				continue
			}
			if config.Ignore(file.Name()) {
				if config.Verbose {
					log.Printf("Ignore %s (in .gitignore)\n", abs+file.Name())
				}
				continue
			}
			if !strings.HasSuffix(file.Name(), ".toml") {
				if config.Verbose {
					log.Printf("Ignore %s (invalid extension)\n", abs+file.Name())
				}
				continue
			}

			// Wrap in anonymous fn to keep open fds low
			e := func() error {
				if config.Verbose {
					log.Printf("Read %s\n", abs+file.Name())
				}

				fd, e := os.Open(filepath.Clean(filepath.Join(abs, file.Name())))
				if e != nil {
					return e
				}

				buf := bufio.NewReader(fd)
				if _, e := toml.NewDecoder(buf).Decode(mem); e != nil {
					if err := fd.Close(); err != nil {
						log.Printf("db.List close: %s", err)
					}
					return e
				}
				title := filepath.Base(fd.Name())
				if e := f(title[0:strings.LastIndex(title, ".")], fd.Name(), path); e != nil {
					if err := fd.Close(); err != nil {
						log.Printf("db.List close: %s", err)
					}
					return e
				}
				i++
				return fd.Close()
			}()
			if e != nil {
				return page, e
			}
		}
	}
	return page, e
}
