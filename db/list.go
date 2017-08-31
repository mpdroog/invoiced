package db

import (
	"io/ioutil"
	"github.com/BurntSushi/toml"
	"fmt"
	"os"
	"bufio"
	"strings"
	"log"
	"github.com/mpdroog/invoiced/config"
	"path/filepath"
)

const MAX_FILES = 1000

type Pagination struct {
	From int // TODO: possible?
	Count int
}

type PaginationHeader struct {
	Total int
}

func List(path []string, p Pagination, mem interface{}, f func(string, string, string) error) (PaginationHeader, error) {
	var page PaginationHeader

	if p.Count > MAX_FILES {
		return page, fmt.Errorf("Pagination.Count exceeds MAX_FILES(%d)", MAX_FILES)
	}

	var pfPath []string
	for _, p := range path {
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		pfPath = append(pfPath, Path+p)
	}

	paths, e := parseWildcards(pfPath)
	if e != nil {
		return page, e
	}

	i := 0
	for _, path := range paths {
		// TODO: pagination??

		files, e := ioutil.ReadDir(path)
		if e != nil {
			return page, e
		}

		abs := path
		for _, file := range files {
			if file.IsDir() {
				if config.Verbose {
					log.Printf("Ignore %s\n", abs + file.Name())
				}
				continue
			}
			if config.Ignore(file.Name()) {
				if config.Verbose {
					log.Printf("Ignore %s\n", abs + file.Name())
				}
				continue
			}

			// Wrap in anonymous fn to keep open fds low
			e := func() error {
				if config.Verbose {
					log.Printf("Read %s\n", abs + file.Name())
				}

				file, e := os.Open(abs + file.Name())
				if e != nil {
					return e
				}
				defer file.Close()

				buf := bufio.NewReader(file)
				if _, e := toml.DecodeReader(buf, mem); e != nil {
					return e
				}
				title := filepath.Base(file.Name())
				if e := f(title[0:strings.LastIndex(title, ".")], file.Name(), path); e != nil {
					return e
				}
				i++
				return nil
			}()
			if e != nil {
				return page, e
			}
		}
	}
	return page, e
}