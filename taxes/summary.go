package taxes

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	//"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/writer"
	"log"
	"strings"
)

type Overview struct {
	Sum  string
	Ex   string
	Tax  string
	EUEx string // TOOODOOO

	NLCompany map[string]string
	EUCompany map[string]string
	Invoices  map[string]string
}

// Create summary for accountant
func Summary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	sum := &Overview{}
	sum.NLCompany = make(map[string]string)
	sum.EUCompany = make(map[string]string)
	sum.Invoices = make(map[string]string)

	e := db.View(func(t *db.Txn) error {
		// invoice
		paths := []string{
			fmt.Sprintf("%s/%s/{all}/sales-invoices-paid", entity, year),
			fmt.Sprintf("%s/%s/{all}/sales-invoices-unpaid", entity, year),
		}
		u := new(invoice.Invoice)
		_, e := t.List(paths, db.Pagination{From: 0, Count: 0}, &u, func(filename, filepath, path string) error {
			var e error
			if config.Verbose {
				log.Printf("Invoice(%s) total=%s ex=%s", u.Meta.Invoiceid, u.Total.Total, u.Total.Ex)
			}
			sum.Sum, e = addValue(sum.Sum, u.Total.Total, 2)
			if e != nil {
				return e
			}

			idname := u.Customer.Name + "-" + u.Customer.Vat
			sum.Invoices[u.Meta.Invoiceid] = u.Total.Total

			if strings.Contains(u.Notes, "VAT Reverse charge") {
				sum.EUEx, e = addValue(sum.EUEx, u.Total.Ex, 2)
				custvat, ok := sum.EUCompany[idname]
				if !ok {
					custvat = "0.00"
				}
				sum.EUCompany[idname], e = addValue(custvat, u.Total.Total, 2)
			} else {
				sum.Ex, e = addValue(sum.Ex, u.Total.Ex, 2)
				custvat, ok := sum.NLCompany[idname]
				if !ok {
					custvat = "0.00"
				}
				sum.NLCompany[idname], e = addValue(custvat, u.Total.Total, 2)
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

	isExcel := false
	if accept := r.Header.Get("Accept"); accept == "application/vnd.ms-excel" {
		isExcel = true
	}
	if r.URL.Query().Get("excel") != "" {
		isExcel = true
	}

	if isExcel {
		// Return Excel-sheet for accountant
		f := excelize.NewFile()
		// overview sheet
		{
			sheet := "Sheet1"
			f.SetCellValue(sheet, "A1", "Revenue")
			f.SetCellValue(sheet, "B1", "RevenueExTax")
			f.SetCellValue(sheet, "C1", "Tax")
			//
			f.SetCellValue(sheet, "A2", sum.Sum)
			f.SetCellValue(sheet, "B2", sum.Ex)
			f.SetCellValue(sheet, "C2", sum.Tax)
		}
		{
			sheet := "Companies"
			f.NewSheet(sheet)
			f.SetCellValue(sheet, "A1", "Company-VAT")
			f.SetCellValue(sheet, "B1", "Revenue")
			pos := 1
			for idname, total := range sum.EUCompany {
				pos++
				f.SetCellValue(sheet, fmt.Sprintf("A%d", pos), idname)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", pos), total)
			}
			for idname, total := range sum.NLCompany {
				pos++
				f.SetCellValue(sheet, fmt.Sprintf("A%d", pos), idname)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", pos), total)
			}
		}
		{
			sheet := "Invoices"
			f.NewSheet(sheet)
			f.SetCellValue(sheet, "A1", "InvoiceID")
			f.SetCellValue(sheet, "B1", "Revenue")
			pos := 1
			for id, total := range sum.Invoices {
				pos++
				f.SetCellValue(sheet, fmt.Sprintf("A%d", pos), id)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", pos), total)
			}
		}

		fname := fmt.Sprintf("%s-%s.xlsx", entity, year)
		w.Header().Set("Content-Type", "application/vnd.ms-excel")
		w.Header().Set("Content-Disposition", "attachment; filename="+fname)
		if _, e := f.WriteTo(w); e != nil {
			log.Printf("summary.excel.WriteTo " + e.Error())
		}
		return
	}

	if e := writer.Encode(w, r, sum); e != nil {
		log.Printf("summary.Summary " + e.Error())
	}
}
