package invoice

import (
	"github.com/boltdb/bolt"
	"encoding/json"
	"fmt"

	"bytes"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"github.com/mpdroog/invoiced/invoice/camt053"
	"github.com/mpdroog/invoiced/config"
	"strings"
	"io"
)

// Parse bankbalance in CAMT053-format
func Balance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
   var buf bytes.Buffer
    file, header, e := r.FormFile("file")
    if e != nil {
    	log.Printf(e.Error())
		http.Error(w, "Failed reading file", 500)
		return
    }
    defer file.Close()
    name := strings.Split(header.Filename, ".")
    if name[1] != "xml" {
    	http.Error(w, "Sorry, not an XML-file", 400)
		return
    }

    io.Copy(&buf, file)
	p, e := camt053.FilterPaymentsReceived(&buf)
	if e != nil {
    	log.Printf(e.Error())
		http.Error(w, "Failed parsing file", 500)
		return
	}

	for _, payment := range p {
		if config.Verbose {
			log.Printf(
				"Parse payments(%s) %sEUR with comment=%s from=%s(%s)",
				payment.Id, payment.Amount, payment.Comment, payment.Name, payment.IBAN,
			)
		}

		ok, e := balanceSetPaid(payment.Comment, payment.Date, payment.Amount)
		if e != nil {
			log.Printf(e.Error())
			http.Error(w, "Failed marking payments as paid", 500)
			return
		}

		if !ok {
			log.Printf("Failed marking payment(%s) as paid", payment.Comment)
		} else if config.Verbose {
			log.Printf("Marked payment(%s) as paid", payment.Comment)
		}
	}
}

func balanceSetPaid(name string, payDate string, amount string) (bool, error) {
	count := 1000
	ok := false
	e := db.Update(func(tx *bolt.Tx) error {
		u := new(Invoice)
		found := false

		b := tx.Bucket([]byte("invoices"))
		c := b.Cursor()
		i := 0
		for k, v := c.Last(); k != nil && i < count; k, v = c.Prev() {
			if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
				return e
			}
			if u.Meta.Invoiceid == name {
				// Found invoice!
				found = true
				break
			}
			i++
		}
		if !found {
			// Skip not found
			log.Printf("Invoice(%s) not found?", name)
			return nil
		}

		if u.Total.Total != amount {
			log.Printf("WARN: Invoice(%s) amounts don't match %sEUR/%sEUR", name, u.Total.Total, amount)
			return nil
		}
		if u.Meta.Status != "FINAL" {
			return fmt.Errorf("invoice.balanceSetPaid can only set paid on FINAL-invoices")
		}
		u.Meta.Paydate = payDate

		// Save any changes..
		buf := new(bytes.Buffer)
		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}
		b2, e := tx.CreateBucketIfNotExists([]byte("invoices-paid"))
		if e != nil {
			return e
		}
		if e := b2.Put([]byte(u.Meta.Conceptid), buf.Bytes()); e != nil {
			return e
		}

		// Delete original
		if e := b.Delete([]byte(u.Meta.Conceptid)); e != nil {
			return e
		}

		ok = true
		return nil
	})
	return ok, e
}