package metrics

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/writer"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"strconv"
	"time"
)

type DashboardMetric struct {
	RevenueTotal string
	RevenueEx    string
	Hours        string
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
	year, e := strconv.Atoi(ps.ByName("year"))
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("metrics.Dashboard failed reading year-arg"), 400)
		return
	}

	m := make(map[string]*DashboardMetric)

	e = db.View(func(t *db.Txn) error {
		// invoice
		paths := []string{
			fmt.Sprintf("%s/%d/{all}/sales-invoices-paid", entity, year),
		}
		u := new(invoice.Invoice)
		_, e := t.List(paths, db.Pagination{From: i, Count: count}, &u, func(filename, filepath, path string) error {
			payDate, e := time.Parse("2006-01-02", u.Meta.Issuedate)
			if e != nil {
				return e
			}
			curyear, month, _ := payDate.Date()
			if curyear != year {
				// ignore, only interested in payments earned in requested year
				return nil
			}
			yearmonth := fmt.Sprintf("%d-%.2d", curyear, month)

			if _, ok := m[yearmonth]; !ok {
				m[yearmonth] = &DashboardMetric{
					RevenueTotal: "0.00",
					RevenueEx:    "0.00",
				}
			}

			if config.Verbose {
				log.Printf("Invoice(date=%s) total=%s ex=%s", yearmonth, u.Total.Total, u.Total.Ex)
			}
			m[yearmonth].RevenueTotal, e = addValue(m[yearmonth].RevenueTotal, u.Total.Total)
			if e != nil {
				return e
			}
			m[yearmonth].RevenueEx, e = addValue(m[yearmonth].RevenueEx, u.Total.Ex)
			if e != nil {
				return e
			}
			return nil
		})
		if e != nil {
			return e
		}

		// hours
		paths = []string{fmt.Sprintf("%s/%d/{all}/hours", entity, year)}
		h := new(hour.Hour)
		_, e = t.List(paths, db.Pagination{From: i, Count: count}, h, func(filename, filepath, path string) error {
			lineDate, e := time.Parse("2006-01-02", h.Lines[0].Day)
			if e != nil {
				return e
			}

			curyear, month, _ := lineDate.Date()
			if curyear != year {
				// ignore, only interested in payments earned in requested year
				return nil
			}
			yearmonth := fmt.Sprintf("%d-%.2d", curyear, month)

			_, ok := m[yearmonth]
			if !ok {
				m[yearmonth] = &DashboardMetric{
					RevenueTotal: "0.00",
					RevenueEx:    "0.00",
				}
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
				log.Printf("Hours(date=%s) hours=%s", yearmonth, hours)
			}
			m[yearmonth].Hours, e = addValue(m[yearmonth].Hours, hours)
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
	if e := writer.Encode(w, r, m); e != nil {
		log.Printf("metrics.Dashboard " + e.Error())
	}
}
