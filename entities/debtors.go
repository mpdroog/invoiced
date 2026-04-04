// Package entities manages company, user, and debtor data.
package entities

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/writer"
)

// Debtor represents a customer or billing entity.
type Debtor struct {
	Name           string
	Street1        string
	Street2        string
	VAT            string
	COC            string
	TAX            string // TODO: validate?
	NoteAdd        string
	BillingEmail   []string
	AccountingCode string // Relation code for accounting software export
}

// Search searches for debtors by name.
func Search(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	args := r.URL.Query()
	query := strings.ToLower(args.Get("query"))

	var debtorList map[string]Debtor
	e := db.View(func(t *db.Txn) error {
		return t.Open(db.DebtorsPath(entity), &debtorList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.Search", e)
		return
	}

	out := []Debtor{}
	for name, debtor := range debtorList {
		name = strings.ToLower(name)
		if strings.Contains(name, query) {
			out = append(out, debtor)
		}
	}
	if e := writer.Encode(w, r, out); e != nil {
		httputil.LogErr("entities.Search", e)
	}
}

// GetDebtor retrieves a debtor by name within a transaction.
func GetDebtor(t *db.Txn, entity, debname string) (*Debtor, error) {
	var debtorList map[string]Debtor
	if e := t.Open(db.DebtorsPath(entity), &debtorList); e != nil {
		return nil, e
	}

	for name, debtor := range debtorList {
		if name == debname {
			return &debtor, nil
		}
	}

	return nil, nil
}
