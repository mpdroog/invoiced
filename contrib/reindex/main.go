// reindex throws away the current index and
// builds a new one
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/blevesearch/bleve"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/invoice"
	"os"
	"path/filepath"
	"strings"
)

var verbose bool

func test(q string) {
	index, e := bleve.Open("./search.db")
	if e != nil {
		panic(e)
	}

	f := bleve.NewTermQuery("rootdev")
	qusr := bleve.NewQueryStringQuery(q)

	req := bleve.NewSearchRequest(bleve.NewConjunctionQuery(f, qusr))
	searchResult, e := index.Search(req)
	if e != nil {
		panic(e)
	}

	fmt.Printf("%+v\n", searchResult)

	if e := index.Close(); e != nil {
		panic(e)
	}
}

func main() {
	flag.BoolVar(&verbose, "v", false, "Verbose-mode (log more)")
	flag.Parse()
	filedb := "../../billingdb/"
	file := "search.db"
	fmt.Printf("Build %s\n", file)

	if !strings.HasSuffix(filedb, "/") {
		filedb += "/"
	}

	mapping := bleve.NewIndexMapping()
	index, e := bleve.New(file, mapping)
	if e != nil {
		panic(e)
	}

	// Scan full dir
	if e := filepath.Walk(filedb, func(path string, f os.FileInfo, err error) error {
		relative := path[len(filedb):len(path)]
		if f.IsDir() {
			if verbose {
				fmt.Printf("Basedir=%s\n", path)
			}
			return nil
		}
		if f.Name() == ".DS_Store" {
			if verbose {
				fmt.Printf("Ignore=%s\n", path)
			}
			return nil
		}

		if strings.Contains(path, "/hours/") {
			if verbose {
				fmt.Printf("Parse hour=%s (%s)\n", path, relative)
			}
			out := new(hour.Hour)
			file, e := os.Open(path)
			if e != nil {
				return e
			}
			defer file.Close()

			buf := bufio.NewReader(file)
			if _, e := toml.DecodeReader(buf, out); e != nil {
				return e
			}

			if e := index.Index(relative, out); e != nil {
				panic(e)
			}
		} else if strings.Contains(path, "/sales-invoices") {
			if verbose {
				fmt.Printf("Parse invoice=%s (%s)\n", path, relative)
			}
			out := new(invoice.Invoice)
			file, e := os.Open(path)
			if e != nil {
				return e
			}
			defer file.Close()

			buf := bufio.NewReader(file)
			if _, e := toml.DecodeReader(buf, out); e != nil {
				return e
			}

			if e := index.Index(relative, out); e != nil {
				panic(e)
			}

		} else {
			fmt.Printf("Unprocessed=%s (idx=%s)\n", path, relative)
		}

		return nil
	}); e != nil {
		panic(e)
	}

	if e := index.Close(); e != nil {
		panic(e)
	}

	//fmt.Printf("Test query\n")
	//test("aws")
}
