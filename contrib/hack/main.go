// hack is a small tool to convert invoiced boltdb
// into the toml-text structure
// go build && rm -rf ./db && ./hack
package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"bufio"
	"bytes"
	"time"
	"strings"

	"encoding/json"
	"github.com/BurntSushi/toml"

	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/invoice"
)

func main() {
	db, e := bolt.Open("/Users/mark/billing.db", 0600, nil)
	if e != nil {
		panic(e)
	}
	defer db.Close()

	// create dirs
	years := []string{"2016", "2017"}
	quarters := []string{"Q1", "Q2", "Q3", "Q4"}
	for _, year := range years {
		for _, q := range quarters {
			base := "./db/rootdev/" + year + "/" + q
			if e := os.MkdirAll(base + "/sales-invoices-unpaid", os.ModePerm); e != nil {
				panic(e)
			}
			if e := os.MkdirAll(base + "/sales-invoices-paid", os.ModePerm); e != nil {
				panic(e)
			}
			if e := os.MkdirAll(base + "/hours", os.ModePerm); e != nil {
				panic(e)
			}
		}
	}

	if e := os.MkdirAll("./db/rootdev/concepts/sales-invoices", os.ModePerm); e != nil {
		panic(e)
	}

	// create hours
	if e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("hours"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			name := string(k)
			name = strings.ToLower(name)
			name = strings.Replace(name, " ", "-", -1)

			u := new(hour.Hour)
			if e := json.NewDecoder(bytes.NewReader(v)).Decode(u); e != nil {
				panic(e)
			}

			day, e := time.Parse("2006-01-02", u.Lines[0].Day) // YYYY-mm-dd
			if e != nil {
				panic(e)
			}
			y := day.Year()
			q := invoice.YearQuarter(day)
			path := fmt.Sprintf("./db/rootdev/%d/Q%d/hours/%s.toml", y, q, name)

			f, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
			if e != nil {
				panic(e)
			}
			defer f.Close()

			buf := bufio.NewWriter(f)
			if e := toml.NewEncoder(buf).Encode(u); e != nil {
				return e
			}
			if e := buf.Flush(); e != nil {
				return e
			}
		}
		return nil
	}); e != nil {
		panic(e)
	}

	// create concept+unpaid invoices
	if e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			name := string(k)
			name = strings.ToLower(name)
			name = strings.Replace(name, " ", "-", -1)

			u := new(invoice.Invoice)
			if e := json.NewDecoder(bytes.NewReader(v)).Decode(u); e != nil {
				panic(e)
			}

			path := "./db/rootdev/concepts/sales-invoices/" + name + ".toml"
			if u.Meta.Issuedate != "" {
				day, e := time.Parse("2006-01-02", u.Meta.Issuedate) // YYYY-mm-dd
				if e != nil {
					panic(e)
				}
				y := day.Year()
				q := invoice.YearQuarter(day)
				path = fmt.Sprintf("./db/rootdev/%d/Q%d/sales-invoices-unpaid/%s.toml", y, q, name)
			}

			f, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
			if e != nil {
				panic(e)
			}
			defer f.Close()

			buf := bufio.NewWriter(f)
			if e := toml.NewEncoder(buf).Encode(u); e != nil {
				return e
			}
			if e := buf.Flush(); e != nil {
				return e
			}
		}
		return nil
	}); e != nil {
		panic(e)
	}

	// create paid invoices
	if e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices-paid"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			name := string(k)
			name = strings.ToLower(name)
			name = strings.Replace(name, " ", "-", -1)

			u := new(invoice.Invoice)
			if e := json.NewDecoder(bytes.NewReader(v)).Decode(u); e != nil {
				panic(e)
			}

			path := "./db/rootdev/concepts/sales-invoices/" + name + ".toml"
			if u.Meta.Issuedate == "" {
				switch u.Meta.Invoiceid {
				case "2016Q2-0001":
					u.Meta.Issuedate = "2016-07-11"
					u.Meta.Duedate = "2016-07-25"
					break
				case "2016Q2-0002":
					u.Meta.Issuedate = "2016-06-23"
					u.Meta.Duedate = "2016-07-07"
					break
				case "2016Q3-0003":
					u.Meta.Issuedate = "2016-07-11"
					u.Meta.Duedate = "2016-07-25"
					break
				case "2016Q3-0004":
					u.Meta.Issuedate = "2016-08-01"
					u.Meta.Duedate = "2016-08-15"
					break
				}
			}

			if u.Meta.Issuedate != "" {
				day, e := time.Parse("2006-01-02", u.Meta.Issuedate) // YYYY-mm-dd
				if e != nil {
					panic(e)
				}
				y := day.Year()
				q := invoice.YearQuarter(day)
				path = fmt.Sprintf("./db/rootdev/%d/Q%d/sales-invoices-paid/%s.toml", y, q, name)
			}

			f, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
			if e != nil {
				panic(e)
			}
			defer f.Close()

			buf := bufio.NewWriter(f)
			if e := toml.NewEncoder(buf).Encode(u); e != nil {
				return e
			}
			if e := buf.Flush(); e != nil {
				return e
			}
		}
		return nil
	}); e != nil {
		panic(e)
	}
}