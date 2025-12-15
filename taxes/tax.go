package taxes

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/writer"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"strings"
)

type Sum struct {
	Ex        string
	Tax       string
	EUEx      string // TOOODOOO
	EUCompany map[string]string
}

func addValue(sum, add string, dec int) (string, error) {
	if sum == "" {
		sum = "0.00"
	}

	s, e := decimal.NewFromString(sum)
	if e != nil {
		return sum, e
	}

	a, e := decimal.NewFromString(add)
	if e != nil {
		return sum, e
	}
	return s.Add(a).StringFixed(int32(dec)), nil
}

func Tax(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	quarter := ps.ByName("quarter")

	sum := &Sum{}
	sum.EUCompany = make(map[string]string)
	audit := ""

	e := db.View(func(t *db.Txn) error {
		// invoice
		paths := []string{
			fmt.Sprintf("%s/%s/%s/sales-invoices-paid", entity, year, quarter),
			fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid", entity, year, quarter),
		}

		u := new(invoice.Invoice)
		_, e := t.List(paths, db.Pagination{From: 0, Count: 0}, &u, func(filename, filepath, path string) error {
			var e error

			if strings.Contains(u.Notes, "Export") {
				// Outside EU means no tax
				audit += fmt.Sprintf("Invoice(%s) Export-Ignored\n", u.Meta.Invoiceid)

			} else if strings.Contains(u.Notes, "VAT Reverse charge") {
				sum.EUEx, e = addValue(sum.EUEx, u.Total.Ex, 2)
				custvat, ok := sum.EUCompany[u.Customer.Vat]
				if !ok {
					custvat = "0.00"
				}

				audit += fmt.Sprintf("Invoice(%s) ICP ex=%s tax=%s\n", u.Meta.Invoiceid, u.Total.Ex, u.Total.Tax)
				sum.EUCompany[u.Customer.Vat], e = addValue(custvat, u.Total.Total, 2)
			} else {
				sum.Ex, e = addValue(sum.Ex, u.Total.Ex, 2)
				audit += fmt.Sprintf("Invoice(%s) NL ex=%s tax=%s\n", u.Meta.Invoiceid, u.Total.Ex, u.Total.Tax)
			}
			if e != nil {
				return e
			}

			sum.Tax, e = addValue(sum.Tax, u.Total.Tax, 2)
			return e
		})
		return e
	})
	if e != nil {
		panic(e)
	}

	if config.Verbose {
		log.Printf("TAX audit:\n%s", audit)
	}

	// Remove decimals (Belastingdienst wants all numbers rounded)
	sum.EUEx, e = addValue(sum.EUEx, "0", 0)
	sum.Ex, e = addValue(sum.Ex, "0", 0)
	sum.Tax, e = addValue(sum.Tax, "0", 0)
	for k, v := range sum.EUCompany {
		sum.EUCompany[k], e = addValue(v, "0", 0)
	}

	if e := writer.Encode(w, r, sum); e != nil {
		log.Printf("taxes.Tax " + e.Error())
	}
}
