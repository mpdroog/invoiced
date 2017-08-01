package metrics

import (
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/hour"
	"github.com/shopspring/decimal"
	"strings"
	"strconv"
)

type DashboardMetric struct {
	RevenueTotal string
	RevenueEx string
	Hours string
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

		b = tx.Bucket([]byte("hours"))
		if b == nil {
			// Empty bucket
			w.Header().Set("Content-Type", "application/json")
			w.Write(nil)
			return nil
		}
		c = b.Cursor()

		i = 0
		// Start from last
		for k, v := c.Last(); k != nil && i < count; k, v = c.Prev() {
			u := new(hour.Hour)
			if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
				return e
			}

			idx := strings.LastIndex(u.Lines[0].Day, "-")
			if idx == -1 {
				log.Printf("WARN: Hour(%s) has no valid issuedate?", u.Lines[0].Day)
				continue
			}
			month := u.Lines[0].Day[0:idx]
			_, ok := m[month]
			if !ok {
				m[month] = &DashboardMetric{}
			}

			hours := "0.00"
			for n := 0; n < len(u.Lines); n++ {
				raw := strconv.FormatFloat(u.Lines[n].Hours, 'f', 0, 64)
				hours, e = addValue(hours, raw)
				if e != nil {
					return e
				}
			}

			log.Printf("Hours(date=%s) hours=%s", month, hours)
			m[month].Hours, e = addValue(m[month].Hours, hours)
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
