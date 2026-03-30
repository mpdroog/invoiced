// Package taxes provides tax calculation and reporting.
package taxes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/writer"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// setCellValue is a helper that logs errors from SetCellValue
func setCellValue(f *excelize.File, sheet, cell string, value interface{}) {
	if err := f.SetCellValue(sheet, cell, value); err != nil {
		log.Printf("taxes.Summary SetCellValue %s:%s: %s", sheet, cell, err)
	}
}

// newSheet is a helper that logs errors from NewSheet
func newSheet(f *excelize.File, name string) {
	if _, err := f.NewSheet(name); err != nil {
		log.Printf("taxes.Summary NewSheet %s: %s", name, err)
	}
}

// InvoiceLine represents a single line item for tax summary export.
type InvoiceLine struct {
	Description  string
	InvoiceID    string
	Debet        string
	Credit       string
	Total        string
	Issuedate    string
	CustomerName string
}

// Overview contains aggregated tax data for a year.
type Overview struct {
	Sum   string
	Ex    string
	Tax   string
	EUEx  string // TOOODOOO
	Hours string

	NLCompany map[string]string
	EUCompany map[string]string
	Invoices  map[string]string

	InvoiceLines map[string]InvoiceLine
}

// Summary creates a tax summary for the accountant.
func Summary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year, e := strconv.Atoi(ps.ByName("year"))
	if e != nil {
		httputil.BadRequest(w, "taxes.Summary year", e)
		return
	}

	// TODO
	relationCodes := map[string]string{
		"XSNews B.V.":          "3",
		"ITS HOSTED":           "4",
		"Money Factory B.V.":   "5",
		"Omniga GmbH & Co. KG": "6",
		"NIMA":                 "7",
		"MyAlo GmbH":           "8",
		"RSP Sales":            "9",
		"Rumah":                "9",
	}

	sum := &Overview{}
	sum.NLCompany = make(map[string]string)
	sum.EUCompany = make(map[string]string)
	sum.Invoices = make(map[string]string)
	sum.InvoiceLines = make(map[string]InvoiceLine)
	sum.Hours = "0"
	sum.EUEx = zeroDecimal

	e = db.View(func(t *db.Txn) error {
		// hours
		paths := []string{fmt.Sprintf("%s/%d/{all}/hours", entity, year)}
		h := new(hour.Hour)
		_, e := t.List(paths, db.Pagination{From: 0, Count: 0}, h, func(_, _, _ string) error {
			hours := zeroDecimal
			var e error
			for n := 0; n < len(h.Lines); n++ {
				raw := strconv.FormatFloat(h.Lines[n].Hours, 'f', 0, 64)
				hours, e = addValue(hours, raw, 0)
				if e != nil {
					return e
				}
			}

			if config.Verbose {
				log.Printf("hours=%s", hours)
			}
			sum.Hours, e = addValue(sum.Hours, hours, 0)
			if e != nil {
				return e
			}
			h.Lines = nil
			return nil
		})
		if e != nil {
			return e
		}

		// invoice
		paths = []string{
			fmt.Sprintf("%s/%d/{all}/sales-invoices-paid", entity, year),
			fmt.Sprintf("%s/%d/{all}/sales-invoices-unpaid", entity, year),
		}
		u := new(invoice.Invoice)
		_, e = t.List(paths, db.Pagination{From: 0, Count: 0}, &u, func(_, _, _ string) error {
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

			switch {
			case strings.Contains(u.Notes, "Export"):
				// Outside EU means no tax
			case strings.Contains(u.Notes, "VAT Reverse charge"):
				sum.EUEx, e = addValue(sum.EUEx, u.Total.Ex, 2)
				if e != nil {
					return e
				}
				custvat, ok := sum.EUCompany[idname]
				if !ok {
					custvat = zeroDecimal
				}
				sum.EUCompany[idname], e = addValue(custvat, u.Total.Total, 2)
				if e != nil {
					return e
				}
			default:
				sum.Ex, e = addValue(sum.Ex, u.Total.Ex, 2)
				if e != nil {
					return e
				}
				custvat, ok := sum.NLCompany[idname]
				if !ok {
					custvat = zeroDecimal
				}
				sum.NLCompany[idname], e = addValue(custvat, u.Total.Total, 2)
				if e != nil {
					return e
				}
			}
			if e != nil {
				return e
			}

			// lines
			for idx, line := range u.Lines {
				// 4647,75/100*21

				extotal, e := decimal.NewFromString(line.Total)
				if e != nil {
					return e
				}
				// TODO: if debtor.TAX == "NL21" {
				tax := extotal.Div(decimal.NewFromFloat(100)).Mul(decimal.NewFromFloat(21))
				total := extotal.Add(tax)

				sum.InvoiceLines[fmt.Sprintf("%s-%d", u.Meta.Invoiceid, idx)] = InvoiceLine{
					Description:  line.Description,
					InvoiceID:    u.Meta.Invoiceid,
					Debet:        line.Total,
					Credit:       tax.StringFixed(2),
					Total:        total.StringFixed(2),
					Issuedate:    u.Meta.Issuedate,
					CustomerName: u.Customer.Name,
				}
			}

			sum.Tax, e = addValue(sum.Tax, u.Total.Tax, 2)
			return e
		})
		return e
	})
	if e != nil {
		httputil.InternalError(w, "taxes.Summary", e)
		return
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
			setCellValue(f, sheet, "A1", "Revenue")
			setCellValue(f, sheet, "B1", "RevenueExTax")
			setCellValue(f, sheet, "C1", "Tax")
			setCellValue(f, sheet, "D1", "Hours")
			//
			setCellValue(f, sheet, "A2", sum.Sum)
			setCellValue(f, sheet, "B2", sum.Ex)
			setCellValue(f, sheet, "C2", sum.Tax)
			setCellValue(f, sheet, "D2", sum.Hours)
		}
		{
			sheet := "Companies"
			newSheet(f, sheet)
			setCellValue(f, sheet, "A1", "Company-VAT")
			setCellValue(f, sheet, "B1", "Revenue")
			pos := 1
			for idname, total := range sum.EUCompany {
				pos++
				setCellValue(f, sheet, fmt.Sprintf("A%d", pos), idname)
				setCellValue(f, sheet, fmt.Sprintf("B%d", pos), total)
			}
			for idname, total := range sum.NLCompany {
				pos++
				setCellValue(f, sheet, fmt.Sprintf("A%d", pos), idname)
				setCellValue(f, sheet, fmt.Sprintf("B%d", pos), total)
			}
		}
		{
			sheet := "Invoices"
			newSheet(f, sheet)
			setCellValue(f, sheet, "A1", "InvoiceID")
			setCellValue(f, sheet, "B1", "AccountingID")
			setCellValue(f, sheet, "C1", "Revenue")
			pos := 1
			for id, total := range sum.Invoices {
				pos++
				acctID := fmt.Sprintf("%d%s", year, strings.Split(id, "-")[1])
				setCellValue(f, sheet, fmt.Sprintf("A%d", pos), id)
				setCellValue(f, sheet, fmt.Sprintf("B%d", pos), acctID)
				setCellValue(f, sheet, fmt.Sprintf("C%d", pos), total)
			}
		}
		{
			sheet := "AccountingSales"
			newSheet(f, sheet)
			setCellValue(f, sheet, "A1", "fldDagboek")
			setCellValue(f, sheet, "B1", "fldBoekingcode")
			setCellValue(f, sheet, "C1", "Datum")
			setCellValue(f, sheet, "D1", "Grootboeknummer")
			setCellValue(f, sheet, "E1", "Debet")
			setCellValue(f, sheet, "F1", "Credit")
			setCellValue(f, sheet, "G1", "ImportBoekingID")
			setCellValue(f, sheet, "H1", "Volgnummer")
			setCellValue(f, sheet, "I1", "Boekstuk")
			setCellValue(f, sheet, "J1", "Omschrijving")
			setCellValue(f, sheet, "K1", "Relatiecode")
			setCellValue(f, sheet, "L1", "Factuurnummer")
			setCellValue(f, sheet, "M1", "Kostenplaatsnummer")

			pos := 1
			for _, line := range sum.InvoiceLines {
				// "2021Q1-0152"
				acctID := fmt.Sprintf("%d%s", year, strings.Split(line.InvoiceID, "-")[1])
				// TODO: hardcoded ledgers
				for _, ledger := range []string{"1300", "1671", "8000"} {
					pos++

					setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "1300")
					setCellValue(f, sheet, fmt.Sprintf("B%d", pos), acctID)
					setCellValue(f, sheet, fmt.Sprintf("C%d", pos), line.Issuedate)
					setCellValue(f, sheet, fmt.Sprintf("D%d", pos), ledger)

					debet := ""
					if ledger == "1300" {
						debet = line.Total
					}
					setCellValue(f, sheet, fmt.Sprintf("E%d", pos), debet)

					credit := ""
					switch ledger {
					case "1671":
						credit = line.Credit
					case "8000":
						credit = line.Debet
					}
					setCellValue(f, sheet, fmt.Sprintf("F%d", pos), credit)

					setCellValue(f, sheet, fmt.Sprintf("G%d", pos), "")
					setCellValue(f, sheet, fmt.Sprintf("H%d", pos), "")
					setCellValue(f, sheet, fmt.Sprintf("I%d", pos), acctID)
					setCellValue(f, sheet, fmt.Sprintf("J%d", pos), line.Description)

					debtorCode := "0"
					if val, ok := relationCodes[line.CustomerName]; ok {
						debtorCode = val
					}
					if debtorCode == "0" {
						fmt.Printf("WARN: Missing debtorCode for %s", line.CustomerName)
					}
					setCellValue(f, sheet, fmt.Sprintf("K%d", pos), debtorCode)

					setCellValue(f, sheet, fmt.Sprintf("L%d", pos), line.InvoiceID)
					setCellValue(f, sheet, fmt.Sprintf("M%d", pos), "")
				}
			}
		}
		{
			sheet := "AccountingPurchases"
			newSheet(f, sheet)
			// TODO: Read from Voorbelasting.txt?
		}

		fname := fmt.Sprintf("%s-%d.xlsx", entity, year)
		w.Header().Set("Content-Type", "application/vnd.ms-excel")
		w.Header().Set("Content-Disposition", "attachment; filename="+fname)
		if _, e := f.WriteTo(w); e != nil {
			httputil.LogErr("taxes.Summary excel.WriteTo", e)
		}
		return
	}

	if e := writer.Encode(w, r, sum); e != nil {
		httputil.LogErr("taxes.Summary encode", e)
	}
}
