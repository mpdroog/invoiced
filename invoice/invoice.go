package invoice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
)

var db *bolt.DB

func Init(d *bolt.DB) error {
	db = d
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("invoices"))
	    if err != nil {
	        return fmt.Errorf("create bucket: %s", err)
	    }
		return nil
	})
}

func Save(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.URL.Query().Get("name")
	data := r.URL.Query().Get("invoice")

	e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		return b.Put([]byte(name), []byte(data))
	})
	if e != nil {
		// todo
	}
}

func Load(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.URL.Query().Get("name")
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
	    v := b.Get([]byte(name))

	    if _, e := w.Write(v); e != nil {
	    	return e
	    }
	    return nil
	})
	if e != nil {
		// todo
	}	
}

func List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("invoices"))
		c := b.Cursor()

		var keys [][]byte
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			keys = append(keys, k)
    	}

    	if e := json.NewEncoder(w).Encode(keys); e != nil {
    		return e
    	}
	    return nil
	})
	if e != nil {
		// todo
	}
}
