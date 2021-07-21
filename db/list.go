package db

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/mpdroog/invoiced/config"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MAX_FILES = 1000
const DEADLINE = "5s"

var deadline time.Duration

type Pagination struct {
	From  int // TODO: possible?
	Count int
}

type PaginationHeader struct {
	Total int
}

func init() {
	var e error
	deadline, e = time.ParseDuration(DEADLINE)
	if e != nil {
		panic(e)
	}
}

func (t *Txn) List(path []string, p Pagination, mem interface{}, f func(string, string, string) error) (PaginationHeader, error) {
	var page PaginationHeader

	if p.Count > MAX_FILES {
		return page, fmt.Errorf("Pagination.Count exceeds MAX_FILES(%d)", MAX_FILES)
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
			return page, fmt.Errorf("Deadline Timeout")
		}
		// TODO: pagination??

		files, e := ioutil.ReadDir(path)
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
				return page, fmt.Errorf("Deadline Timeout")
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

				file, e := os.Open(abs + file.Name())
				if e != nil {
					return e
				}

				buf := bufio.NewReader(file)
				if _, e := toml.DecodeReader(buf, mem); e != nil {
					file.Close() /* ignore err, write err takes precedence */
					return e
				}
				title := filepath.Base(file.Name())
				if e := f(title[0:strings.LastIndex(title, ".")], file.Name(), path); e != nil {
					file.Close() /* ignore err, write err takes precedence */
					return e
				}
				i++
				return file.Close()
			}()
			if e != nil {
				return page, e
			}
		}
	}
	return page, e
}
