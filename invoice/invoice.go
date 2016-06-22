package invoice

import (
	"bytes"
	"log"
	"encoding/json"
	"net/http"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"time"
)

var db *bolt.DB

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

	buf := new(bytes.Buffer)
	if e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		if u.Meta.Invoiceid == "" {
			// Create one
			now := time.Now()
			var quarter string
			switch now.Month() {
				case time.January:
				case time.February:
				case time.March:
					quarter = "Q1"
				case time.April:
				case time.May:
				case time.June:
					quarter = "Q2"
				case time.July:
				case time.August:
				case time.September:
					quarter = "Q3"
				case time.October:
				case time.November:
				case time.December:
					quarter = "Q4"
			}
			idx, e := b.NextSequence()
			if e != nil {
				return e
			}

			u.Meta.Invoiceid = fmt.Sprintf("%d%s-%04d", now.Year(), quarter, idx)
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
	    _, e := w.Write(v);
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
