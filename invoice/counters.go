package invoice

import (
	"fmt"
	"github.com/mpdroog/invoiced/db"
)

// Counter holds sequential counters for invoice IDs.
type Counter struct {
	InvoiceID uint64
}

// NextInvoiceID returns and increments the next invoice ID for an entity.
func NextInvoiceID(entity string, t *db.Txn) (uint64, error) {
	c := &Counter{}
	path := fmt.Sprintf("%s/counters.toml", entity)
	if e := t.Open(path, c); e != nil {
		return 0, e
	}

	c.InvoiceID++
	if e := t.Save(path, false, c); e != nil {
		return 0, e
	}

	return c.InvoiceID, nil
}
