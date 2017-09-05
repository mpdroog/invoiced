package metrics

import (
	//"bytes"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/config"
	"github.com/shopspring/decimal"
	"strings"
	"strconv"
	"fmt"
)

type DashboardMetric struct {
	RevenueTotal string
	RevenueEx string
	Hours string
}

func addValue(sum, add string) (string, error) {
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
	return s.Add(a).StringFixed(2), nil
}

func Dashboard(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// range
	i := 0
	count := 999 // TODO: remove invoice limit?
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	m := make(map[string]*DashboardMetric)

	e := db.View(func(t *db.Txn) error {
		// invoice
		paths := []string{fmt.Sprintf("%s/%s/{all}/sales-invoices-paid", entity, year)}
		u := new(invoice.Invoice)
		_, e := t.List(paths, db.Pagination{From:i, Count:count}, &u, func(filename, filepath, path string) error {
			idx := strings.LastIndex(u.Meta.Issuedate, "-")
			if idx == -1 {
				log.Printf("WARN: Invoice(%s) has no valid issuedate?", u.Meta.Invoiceid)
				return nil
			}
			month := u.Meta.Issuedate[0:idx]
			_, ok := m[month]
			if !ok {
				m[month] = &DashboardMetric{}
			}

			var e error
			if config.Verbose {
				log.Printf("Invoice(date=%s) total=%s ex=%s", month, u.Total.Total, u.Total.Ex)
			}
			m[month].RevenueTotal, e = addValue(m[month].RevenueTotal, u.Total.Total)
			if e != nil {
				return e
			}
			m[month].RevenueEx, e = addValue(m[month].RevenueEx, u.Total.Ex)
			if e != nil {
				return e
			}
			return nil
		})
		if e != nil {
			return e
		}

		// hours
		paths = []string{fmt.Sprintf("%s/%s/{all}/hours", entity, year)}
		h := new(hour.Hour)
		_, e = t.List(paths, db.Pagination{From:i, Count:count}, h, func(filename, filepath, path string) error {
			idx := strings.LastIndex(h.Lines[0].Day, "-")
			if idx == -1 {
				log.Printf("WARN: Hour(%s) has no valid issuedate?", h.Lines[0].Day)
				return nil
			}
			month := h.Lines[0].Day[0:idx]
			_, ok := m[month]
			if !ok {
				m[month] = &DashboardMetric{}
			}

			hours := "0.00"
			for n := 0; n < len(h.Lines); n++ {
				raw := strconv.FormatFloat(h.Lines[n].Hours, 'f', 0, 64)
				hours, e = addValue(hours, raw)
				if e != nil {
					return e
				}
			}

			if config.Verbose {
				log.Printf("Hours(date=%s) hours=%s", month, hours)
			}
			m[month].Hours, e = addValue(m[month].Hours, hours)
			if e != nil {
				return e
			}
			h.Lines = nil
			return nil
		})
		return e
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("metrics.Dashboard failed scanning disk"), 400)
		return
	}
	if config.Verbose {
		log.Printf("metrics.Dashboard count=%d", len(m))
	}
	//w.Header().Set("X-Pagination-Total", string(p.Total))
	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(m); e != nil {
		log.Printf(e.Error())
		return
	}
}
