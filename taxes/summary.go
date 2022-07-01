package taxes

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/writer"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type InvoiceLine struct {
	Description  string
	InvoiceId    string
	Debet        string
	Credit       string
	Total        string
	Issuedate    string
	CustomerName string
}

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

// Create summary for accountant
func Summary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year, e := strconv.Atoi(ps.ByName("year"))
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("taxes.Summary failed reading year-arg"), 400)
		return
	}

	// TODO
	relationCodes := map[string]string{
		"XSNews B.V.":          "3",
		"ITS HOSTED":           "4",
		"Money Factory B.V.":   "5",
		"Omniga GmbH & Co. KG": "6",
		"NIMA":                 "7",
	}

	sum := &Overview{}
	sum.NLCompany = make(map[string]string)
	sum.EUCompany = make(map[string]string)
	sum.Invoices = make(map[string]string)
	sum.InvoiceLines = make(map[string]InvoiceLine)
	sum.Hours = "0"
	sum.EUEx = "0.00"

	e = db.View(func(t *db.Txn) error {
		// hours
		paths := []string{fmt.Sprintf("%s/%d/{all}/hours", entity, year)}
		h := new(hour.Hour)
		_, e := t.List(paths, db.Pagination{From: 0, Count: 0}, h, func(filename, filepath, path string) error {
			hours := "0.00"
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
		_, e = t.List(paths, db.Pagination{From: 0, Count: 0}, &u, func(filename, filepath, path string) error {
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

			// lines
			for idx, line := range u.Lines {
				// 4647,75/100*21

				extotal, e := decimal.NewFromString(line.Total)
				if e != nil {
					return e
				}
				tax := decimal.NewFromFloat(0)
				// TODO: if debtor.TAX == "NL21" {
				tax = extotal.Div(decimal.NewFromFloat(100)).Mul(decimal.NewFromFloat(21))
				total := extotal.Add(tax)

				sum.InvoiceLines[fmt.Sprintf("%s-%d", u.Meta.Invoiceid, idx)] = InvoiceLine{
					Description:  line.Description,
					InvoiceId:    u.Meta.Invoiceid,
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
			f.SetCellValue(sheet, "D1", "Hours")
			//
			f.SetCellValue(sheet, "A2", sum.Sum)
			f.SetCellValue(sheet, "B2", sum.Ex)
			f.SetCellValue(sheet, "C2", sum.Tax)
			f.SetCellValue(sheet, "D2", sum.Hours)
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
			f.SetCellValue(sheet, "B1", "AccountingID")
			f.SetCellValue(sheet, "C1", "Revenue")
			pos := 1
			for id, total := range sum.Invoices {
				pos++
				acctId := fmt.Sprintf("%d%s", year, strings.Split(id, "-")[1])
				f.SetCellValue(sheet, fmt.Sprintf("A%d", pos), id)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", pos), acctId)
				f.SetCellValue(sheet, fmt.Sprintf("C%d", pos), total)
			}
		}
		{
			sheet := "AccountingSales"
			f.NewSheet(sheet)
			f.SetCellValue(sheet, "A1", "fldDagboek")
			f.SetCellValue(sheet, "B1", "fldBoekingcode")
			f.SetCellValue(sheet, "C1", "Datum")
			f.SetCellValue(sheet, "D1", "Grootboeknummer")
			f.SetCellValue(sheet, "E1", "Debet")
			f.SetCellValue(sheet, "F1", "Credit")
			f.SetCellValue(sheet, "G1", "ImportBoekingID")
			f.SetCellValue(sheet, "H1", "Volgnummer")
			f.SetCellValue(sheet, "I1", "Boekstuk")
			f.SetCellValue(sheet, "J1", "Omschrijving")
			f.SetCellValue(sheet, "K1", "Relatiecode")
			f.SetCellValue(sheet, "L1", "Factuurnummer")
			f.SetCellValue(sheet, "M1", "Kostenplaatsnummer")

			pos := 1
			for _, line := range sum.InvoiceLines {
				// "2021Q1-0152"
				acctId := fmt.Sprintf("%d%s", year, strings.Split(line.InvoiceId, "-")[1])
				// TODO: hardcoded ledgers
				for _, ledger := range []string{"1300", "1671", "8000"} {
					pos++

					f.SetCellValue(sheet, fmt.Sprintf("A%d", pos), "1300")
					f.SetCellValue(sheet, fmt.Sprintf("B%d", pos), acctId)
					f.SetCellValue(sheet, fmt.Sprintf("C%d", pos), line.Issuedate)
					f.SetCellValue(sheet, fmt.Sprintf("D%d", pos), ledger)

					debet := ""
					if ledger == "1300" {
						debet = line.Total
					}
					f.SetCellValue(sheet, fmt.Sprintf("E%d", pos), debet)

					credit := ""
					if ledger == "1671" {
						credit = line.Credit
					} else if ledger == "8000" {
						credit = line.Debet
					}
					f.SetCellValue(sheet, fmt.Sprintf("F%d", pos), credit)

					f.SetCellValue(sheet, fmt.Sprintf("G%d", pos), "")
					f.SetCellValue(sheet, fmt.Sprintf("H%d", pos), "")
					f.SetCellValue(sheet, fmt.Sprintf("I%d", pos), acctId)
					f.SetCellValue(sheet, fmt.Sprintf("J%d", pos), line.Description)

					debtorCode := "0"
					if val, ok := relationCodes[line.CustomerName]; ok {
						debtorCode = val
					}
					f.SetCellValue(sheet, fmt.Sprintf("K%d", pos), debtorCode)

					f.SetCellValue(sheet, fmt.Sprintf("L%d", pos), acctId)
					f.SetCellValue(sheet, fmt.Sprintf("M%d", pos), "")
				}
			}
		}
		{
			sheet := "AccountingPurchases"
			f.NewSheet(sheet)
			// TODO: Read from Voorbelasting.txt?
		}

		fname := fmt.Sprintf("%s-%d.xlsx", entity, year)
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
