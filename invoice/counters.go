package invoice

import (
	"github.com/mpdroog/invoiced/db"
	"fmt"
)

type Counter struct {
	InvoiceID uint64
}

func NextInvoiceID(entityYear string, t *db.Txn) (uint64, error) {
	c := &Counter{}
	path := fmt.Sprintf("%s/counters.toml", entityYear)
	if e := t.Open(path, c); e != nil {
		return 0, e
	}

	c.InvoiceID++
	if e := t.Save(path, c); e != nil {
		return 0, e
	}

	return c.InvoiceID, nil
}