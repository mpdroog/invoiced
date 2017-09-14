package entities

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/writer"
	"strings"
	"fmt"
	"log"
)

type Debtor struct {
	Name string
	Street1 string
	Street2 string
	VAT string
	COC string
	TAX string // TODO: validate?
	NoteAdd string
	BillingEmail []string
}

func Search(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	args := r.URL.Query()
	query := strings.ToLower(args.Get("query"))

	var debtorList map[string]Debtor
	e := db.View(func(t *db.Txn) error {
		return t.Open(fmt.Sprintf("%s/debtors.toml", entity), &debtorList)
	})
	if e != nil {
		log.Printf("entities.Search e=" + e.Error())
		http.Error(w, "Failed reading debtors", 500)
		return
	}

	var out []Debtor
	for name, debtor := range debtorList {
		name = strings.ToLower(name)
		if strings.Contains(name, query) {
			log.Printf("Contains %s/%s\n", query, name)
			out = append(out, debtor)
		}
	}
	if e := writer.Encode(w, r, out); e != nil {
		log.Printf("entities.Search " + e.Error())
	}
}