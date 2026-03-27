package taxes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/idx"
	"github.com/mpdroog/invoiced/writer"
	"github.com/shopspring/decimal"
)

type Sum struct {
	Ex        string // Sum revenue of NL invoices
	Tax       string // Tax to pay
	EUEx      string // Sum revenue of EU invoices
	EUCompany map[string]string // Tax per EU company for ICP

	ExWorld   string // Sum revenue of world invoices
	ExRevenue string // Sum revenue of everything
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

	if idx.DB == nil {
		http.Error(w, "Index not initialized", 500)
		return
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		http.Error(w, "Invalid year", 400)
		return
	}
	quarterInt, err := strconv.Atoi(quarter[1:]) // "Q1" -> 1
	if err != nil {
		http.Error(w, "Invalid quarter", 400)
		return
	}

	idxSum, audit, err := idx.GetQuarterTaxSummary(entity, yearInt, quarterInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %s", err.Error()), 500)
		return
	}

	sum := &Sum{
		Ex:        idxSum.Ex,
		Tax:       idxSum.Tax,
		EUEx:      idxSum.EUEx,
		EUCompany: idxSum.EUCompany,
		ExWorld:   idxSum.ExWorld,
		ExRevenue: idxSum.ExRevenue,
	}

	if config.Verbose {
		auditStr := ""
		for _, a := range audit {
			auditStr += fmt.Sprintf("Invoice(%s) %s ex=%s tax=%s\n",
				a.InvoiceID, a.TaxCategory, a.TotalEx, a.TotalTax)
		}
		log.Printf("TAX audit:\n%s", auditStr)
	}

	// Remove decimals (Belastingdienst wants all numbers rounded)
	var e error
	sum.EUEx, e = addValue(sum.EUEx, "0", 0)
	if e != nil {
		http.Error(w, fmt.Sprintf("EUEx rounding failed: %s", e.Error()), 500)
		return
	}
	sum.Ex, e = addValue(sum.Ex, "0", 0)
	if e != nil {
		http.Error(w, fmt.Sprintf("Ex rounding failed: %s", e.Error()), 500)
		return
	}
	sum.Tax, e = addValue(sum.Tax, "0", 0)
	if e != nil {
		http.Error(w, fmt.Sprintf("Tax rounding failed: %s", e.Error()), 500)
		return
	}
	for k, v := range sum.EUCompany {
		sum.EUCompany[k], e = addValue(v, "0", 0)
		if e != nil {
			http.Error(w, fmt.Sprintf("EUCompany[%s] rounding failed: %s", k, e.Error()), 500)
			return
		}
	}

	if e := writer.Encode(w, r, sum); e != nil {
		http.Error(w, fmt.Sprintf("Encode failed: %s", e.Error()), 500)
		return
	}
}
