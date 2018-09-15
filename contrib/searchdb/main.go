package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mpdroog/invoiced/invoice"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var verbose bool

func main() {
	flag.BoolVar(&verbose, "v", false, "Verbose-mode (log more)")
	flag.Parse()

	db, e := sql.Open("sqlite3", "file:foo.db?cache=shared")
	if e != nil {
		panic(e)
	}
	defer func() {
		if e := db.Close(); e != nil {
			panic(e)
		}
	}()

	sql, e := ioutil.ReadFile("main.sql")
	if e != nil {
		panic(e)
	}

	// prep
	_, e = db.Exec(string(sql))
	if e != nil {
		panic(e)
	}

	stmtInv, e := db.Prepare(`INSERT INTO invoices(total_ex, total_tax, total_sum, concept_id, invoice_id, customer_name) values(?,?,?,?,?,?)`)
	if e != nil {
		panic(e)
	}
	stmtInvLine, e := db.Prepare(`INSERT INTO invoice_lines(invoices_id, description, quantity, price, total) values(?,?,?,?,?)`)
	if e != nil {
		panic(e)
	}

	filedb := "../../billingdb-live/"
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

		if strings.Contains(path, "/sales-invoices") {
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

			res, e := stmtInv.Exec(out.Total.Ex, out.Total.Tax, out.Total.Total, out.Meta.Conceptid, out.Meta.Invoiceid, out.Customer.Name)
			if e != nil {
				return e
			}

			id, e := res.LastInsertId()
			if e != nil {
				return e
			}

			for _, line := range out.Lines {
				_, e = stmtInvLine.Exec(id, line.Description, line.Quantity, line.Price, line.Total)
				if e != nil {
					return e
				}
			}

		} else {
			fmt.Printf("Unprocessed=%s (idx=%s)\n", path, relative)
		}

		return nil
	}); e != nil {
		panic(e)
	}
}
