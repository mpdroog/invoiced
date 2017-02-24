package migrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/mpdroog/invoiced/invoice"
	"log"
	"strconv"
)

const LATEST = 1

func conv0(tx *bolt.Tx) error {
	b := tx.Bucket([]byte("invoices"))
	tmp, e := tx.CreateBucketIfNotExists([]byte("invoices-tmp"))
	if e != nil {
		return e
	}

	// 1) Set missing CONCEPT-ID in invoices-tmp
	idx := 0
	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		fmt.Printf("key=%s, value=%s\n", k, v)

		u := new(invoice.Invoice)
		if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
			return e
		}

		idx++
		if len(u.Meta.Conceptid) == 0 {
			u.Meta.Conceptid = fmt.Sprintf("CONCEPT-%d", idx)
		}
		u.Meta.Status = "FINAL"

		// Save any changes..
		buf := new(bytes.Buffer)
		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}
		fmt.Printf("Write key=%s with val=%s\n", u.Meta.Conceptid, buf.Bytes())
		if e := tmp.Put([]byte(u.Meta.Conceptid), buf.Bytes()); e != nil {
			return e
		}
	}

	// 2) Re-create invoices-bucket
	if e := tx.DeleteBucket([]byte("invoices")); e != nil {
		return e
	}
	b, e = tx.CreateBucketIfNotExists([]byte("invoices"))
	if e != nil {
		return e
	}
	if e := b.SetSequence(uint64(idx)); e != nil {
		return e
	}

	// 3) Convert from -tmp bucket to default bucket
	c = tmp.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if e := b.Put(k, v); e != nil {
			return e
		}
	}
	return tx.DeleteBucket([]byte("invoices-tmp"))
}

// Convert BoltDB from old to new version
func Convert(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte("config"))
		if e != nil {
			return e
		}
		v := b.Get([]byte("version"))
		if v == nil {
			// new db
			log.Printf("Set version to %d\n", LATEST)
			return b.Put([]byte("version"), []byte(strconv.Itoa(LATEST)))
		}

		version, e := strconv.Atoi(string(v))
		if e != nil {
			return e
		}
		if version == LATEST {
			log.Printf("Running latest version %d\n", LATEST)
			return nil
		}

		return fmt.Errorf("Unsupported version=%s", v)
	})
}
