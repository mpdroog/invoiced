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

func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to delete", 400)
		return
	}
	if e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		v := b.Get([]byte(name))
		if v == nil {
			http.Error(w, "invoice.Delete no such name", http.StatusNotFound)
			return nil
		}

		buf := bytes.NewBuffer(v)
		u := new(Invoice)
		if e := json.NewDecoder(buf).Decode(u); e != nil {
			return e
		}
		if u.Meta.Status == "FINAL" {
			// Cannot delete finalized invoices
			http.Error(w, "invoice.Delete cannot delete finalized invoice", 400)
			return nil
		}

		if e := b.Delete([]byte(name)); e != nil {
			return e
		}

		w.Header().Set("Content-Type", "application/json")
		if _, e := w.Write(buf.Bytes()); e != nil {
			return e
		}
		return nil
	}); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Delete fail", http.StatusInternalServerError)
	}
}

// Lock invoice for changes and set invoiceid
func Finalize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	log.Printf("invoice.Finalize with conceptid=%s", name)
	e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		v := b.Get([]byte(name))
		if v == nil {
			http.Error(w, "invoice.Finalize no such name", http.StatusNotFound)
			return nil
		}

		u := new(Invoice)
		if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
			return e
		}

		if len(u.Meta.Issuedate) == 0 {
			u.Meta.Issuedate = time.Now().Format("2006-01-02")
		}

		if u.Meta.Invoiceid == "" {
			// Create invoiceid
			idx, e := b.NextSequence()
			if e != nil {
				return e
			}

			u.Meta.Invoiceid = createInvoiceId(time.Now(), idx)
			log.Printf("invoice.Finalize create conceptId=%s invoiceId=%s", u.Meta.Conceptid, u.Meta.Invoiceid)
		}
		u.Meta.Status = "FINAL"

		// Save any changes..
		buf := new(bytes.Buffer)
		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}
		if e := b.Put([]byte(u.Meta.Conceptid), buf.Bytes()); e != nil {
			return e
		}

		w.Header().Set("Content-Type", "application/json")
		_, e := w.Write(buf.Bytes())
		return e
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Finalize fail", http.StatusInternalServerError)
	}
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
		http.Error(w, "invoice.Save failed to decode input", 400)
		return
	}
	if e := validator.Validate(u); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Save failed validate=%s", e), 400)
		return
	}

	buf := new(bytes.Buffer)
	if e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		// TODO: Check if current entry is FINAL, if so prevent!

		if u.Meta.Conceptid == "" {
			u.Meta.Conceptid = fmt.Sprintf("CONCEPT-%s", randStringBytesRmndr(6))
			log.Printf("invoice.Save create conceptId=%s", u.Meta.Conceptid)
		} else {
			log.Printf("invoice.Save update conceptId=%s", u.Meta.Conceptid)			
		}
		u.Meta.Status = "CONCEPT"

		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}
		return b.Put([]byte(u.Meta.Conceptid), buf.Bytes())
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
	log.Printf("invoice.Load with conceptid=%s", name)
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

		keys := make(map[string]*Invoice)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//keys = append(keys, string(k))
			u := new(Invoice)
			if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
				return e
			}

			keys[string(k)] = u
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
