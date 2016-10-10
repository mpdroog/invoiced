package invoice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"time"
	"gopkg.in/validator.v2"
	"math/rand"
)

var db *bolt.DB
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
    rand.Seed(time.Now().UnixNano())
}

// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func randStringBytesRmndr(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
    }
    return string(b)
}

func Init(d *bolt.DB) error {
	db = d
	return db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte("invoices"))
		return e
	})
}

func Save(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// IF POST create InvoiceID = 2016Q3-0001
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	u := new(Invoice)
	if e := json.NewDecoder(r.Body).Decode(u); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Save fail", http.StatusInternalServerError)
		return
	}
	if e := validator.Validate(u); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Save err, failed validate=%s", e), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	if e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		if u.Meta.Invoiceid == "" {
			// Create one
			idx, e := b.NextSequence()
			if e != nil {
				return e
			}

			u.Meta.Invoiceid = createInvoiceId(time.Now(), idx)
			log.Printf("invoice.Save create invoiceId=%s", u.Meta.Invoiceid)
		} else {
			log.Printf("invoice.Save update invoiceId=%s", u.Meta.Invoiceid)
		}

		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}

		return b.Put([]byte(u.Meta.Invoiceid), buf.Bytes())
	}); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Save fail", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, e := w.Write(buf.Bytes()); e != nil {
		log.Printf(e.Error())
	}
}

func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	log.Printf("invoice.Load with id=%s", name)
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		v := b.Get([]byte(name))
		if v == nil {
			http.Error(w, "invoice.Load no such name", http.StatusNotFound)
			return nil
		}

		w.Header().Set("Content-Type", "application/json")
		_, e := w.Write(v)
		return e
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Save fail", http.StatusInternalServerError)
	}
}

func List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		c := b.Cursor()

		var keys []string
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		log.Printf("invoice.List count=%d", len(keys))

		w.Header().Set("Content-Type", "application/json")
		if e := json.NewEncoder(w).Encode(keys); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Save fail", http.StatusInternalServerError)
	}
}

func Pdf(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	log.Printf("invoice.Pdf with id=%s", name)
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		v := b.Get([]byte(name))
		if v == nil {
			return fmt.Errorf("No such invoice name")
		}

		u := new(Invoice)
		if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
			return e
		}

		//if len(u.Meta.Issuedate) == 0 {
		u.Meta.Issuedate = time.Now().Format("2006-01-02")
		//}
		// TODO: Update in DB

		f, e := pdf(u)
		if e != nil {
			return e
		}

		w.Header().Set("Content-Type", "application/pdf")
		return f.Output(w)
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Pdf fail", http.StatusInternalServerError)
	}
}

func Credit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	log.Printf("invoice.Credit with id=%s", name)
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		v := b.Get([]byte(name))
		if v == nil {
			return fmt.Errorf("No such invoice name")
		}

		u := new(Invoice)
		if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
			return e
		}

		u.Meta.Issuedate = time.Now().Format("2006-01-02")
		// TODO: Update in DB

		f, e := pdf(u)
		if e != nil {
			return e
		}

		w.Header().Set("Content-Type", "application/pdf")
		return f.Output(w)
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Pdf fail", http.StatusInternalServerError)
	}
}
