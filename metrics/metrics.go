package metrics

import (
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/shopspring/decimal"
	"strings"
)

type DashboardMetric struct {
	RevenueTotal string
	RevenueEx string
}

var db *bolt.DB

func Init(d *bolt.DB) error {
	db = d
	return nil
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

func Dashboard(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// range
	i := 0
	count := 1000 // TODO: remove invoice limit?

	m := make(map[string]*DashboardMetric)

	e := db.View(func(tx *bolt.Tx) error {
		var e error
		b := tx.Bucket([]byte("invoices-paid"))
		if b == nil {
			// Empty bucket
			w.Header().Set("Content-Type", "application/json")
			w.Write(nil)
			return nil
		}
		c := b.Cursor()

		// Start from last
		for k, v := c.Last(); k != nil && i < count; k, v = c.Prev() {
			// TODO: Move struct to outside of struct?
			u := new(invoice.Invoice)
			if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
				return e
			}
			idx := strings.LastIndex(u.Meta.Issuedate, "-")
			if idx == -1 {
				log.Printf("WARN: Invoice(%s) has no valid issuedate?", u.Meta.Invoiceid)
				continue
			}
			month := u.Meta.Issuedate[0:idx]
			_, ok := m[month]
			if !ok {
				m[month] = &DashboardMetric{}
			}

			log.Printf("Invoice(date=%s) total=%s ex=%s", month, u.Total.Total, u.Total.Ex)
			m[month].RevenueTotal, e = addValue(m[month].RevenueTotal, u.Total.Total)
			if e != nil {
				return e
			}
			m[month].RevenueEx, e = addValue(m[month].RevenueEx, u.Total.Ex)
			if e != nil {
				return e
			}

			i++
		}

		w.Header().Set("Content-Type", "application/json")
		if e := json.NewEncoder(w).Encode(m); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "metrics.Dashboard fail", http.StatusInternalServerError)
	}
}
