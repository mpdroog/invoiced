package hour

import (
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

var db *bolt.DB

func Init(d *bolt.DB) error {
	db = d
	return db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte("hours"))
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
		b := tx.Bucket([]byte("hours"))
		return b.Delete([]byte(name))
	}); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Delete fail", http.StatusInternalServerError)
	}
}

func Save(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	u := new(Hour)
	if e := json.NewDecoder(r.Body).Decode(u); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Save fail", http.StatusInternalServerError)
		return
	}
	if u.Name == "" {
		http.Error(w, "invoice.Save err, no Name given", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	if e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("hours"))
		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}

		return b.Put([]byte(u.Name), buf.Bytes())
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
		b := tx.Bucket([]byte("hours"))
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
		b := tx.Bucket([]byte("hours"))
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
