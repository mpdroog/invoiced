// hack is a small tool to convert invoiced boltdb
// into the toml-text structure
// go build && rm -rf ./db && ./hack
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"math"
	"os"
	"strings"
	"time"

	"encoding/json"
	"github.com/BurntSushi/toml"

	revfs "github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/utils"
	"io/ioutil"
	"path/filepath"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	//gitconfig "github.com/go-git/go-git/v6/config"
)

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

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
			if e := os.MkdirAll(base+"/sales-invoices-unpaid", os.ModePerm); e != nil {
				panic(e)
			}
			if e := os.MkdirAll(base+"/sales-invoices-paid", os.ModePerm); e != nil {
				panic(e)
			}
			if e := os.MkdirAll(base+"/hours", os.ModePerm); e != nil {
				panic(e)
			}
		}
	}

	for _, year := range years {
		if e := os.MkdirAll("./db/rootdev/"+year+"/concepts/sales-invoices", os.ModePerm); e != nil {
			panic(e)
		}
		if e := os.MkdirAll("./db/rootdev/"+year+"/concepts/hours", os.ModePerm); e != nil {
			panic(e)
		}
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
			for key, val := range u.Lines {
				u.Lines[key].Hours = toFixed(val.Hours, 2)
			}

			day, e := time.Parse("2006-01-02", u.Lines[0].Day) // YYYY-mm-dd
			if e != nil {
				panic(e)
			}
			y := day.Year()
			q := utils.YearQuarter(day)
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

			path := "./db/rootdev/2017/concepts/sales-invoices/" + name + ".toml"
			if u.Meta.Issuedate != "" {
				day, e := time.Parse("2006-01-02", u.Meta.Issuedate) // YYYY-mm-dd
				if e != nil {
					panic(e)
				}
				y := day.Year()
				q := utils.YearQuarter(day)
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

			path := "./db/rootdev/2017/concepts/sales-invoices/" + name + ".toml"
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
				q := utils.YearQuarter(day)
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

	d := []byte(`IV="a5FbH/LiRl5XtLOLvL0MSrJLU8BxP8Ao"
Version=1

[company.rootdev]
Name="RootDev"
COC="65898621"
VAT="NL067931959B01"
IBAN="NL17 RABO 0310 0295 97"
BIC="RABONL2U"
Salt="Ab/tViXI1AvpKxaYvuJ5VphRYCb/gTqH"

[[user]]
Email="rootdev@gmail.com"
Hash=""
Company=["rootdev"]
Name="Mark Droog"
Address1="Dorpsstraat 236a"
Address2="1713HP Obdam"
`)
	if e := ioutil.WriteFile("./db/entities.toml", d, 0755); e != nil {
		panic(e)
	}

	// Now revision it all
	if e := revfs.Init("db"); e != nil {
		panic(e)
	}
	Repo := revfs.Repo

	tree, e := Repo.Worktree()
	if e != nil {
		panic(e)
	}
	e = filepath.Walk("db", func(path string, f os.FileInfo, err error) error {
		fmt.Printf("Read %s\n", path)
		if strings.HasPrefix(path, "db/.git/") {
			return nil
		}
		if f.IsDir() {
			return nil
		}

		fmt.Printf("Add %s\n", path[len("db/"):])
		_, e := tree.Add(path[len("db/"):])
		return e
	})
	if e != nil {
		panic(e)
	}
	opts := &git.CommitOptions{Author: &object.Signature{Name: "root", Email: "support@boekhoud.cloud", When: time.Now()}}
	if e := opts.Validate(Repo); e != nil {
		panic(e)
	}
	if _, e := tree.Commit("Initial commit", opts); e != nil {
		panic(e)
	}
}
