// Package taxes provides tax calculation and reporting.
package taxes

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/idx"
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

// setCellFloat sets a cell to a float value with 2 decimal places
func setCellFloat(f *excelize.File, sheet, cell string, value float64) {
	if err := f.SetCellFloat(sheet, cell, value, 2, 64); err != nil {
		log.Printf("taxes.Summary SetCellFloat %s:%s: %s", sheet, cell, err)
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

	// Optional quarter filter (0 = all quarters)
	quarter := 0
	if q := r.URL.Query().Get("quarter"); q != "" {
		if qInt, err := strconv.Atoi(q); err == nil && qInt >= 1 && qInt <= 4 {
			quarter = qInt
		}
	}

	// Excel export via ?excel=1
	if r.URL.Query().Get("excel") != "" && idx.DB != nil {
		writeAccountingExcel(w, entity, year, quarter)
		return
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
			for _, line := range h.Lines {
				raw := strconv.FormatFloat(line.Hours, 'f', 0, 64)
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

	if e := writer.Encode(w, r, sum); e != nil {
		httputil.LogErr("taxes.Summary encode", e)
	}
}

// writeAccountingExcel generates an improved Excel file for the accountant using SQLite index
func writeAccountingExcel(w http.ResponseWriter, entity string, year, quarter int) {
	export, err := idx.GetAccountingExport(entity, year, quarter)
	if err != nil {
		httputil.InternalError(w, "taxes.Summary idx.GetAccountingExport", err)
		return
	}

	f := excelize.NewFile()

	// Sheet 1: Overview
	{
		sheet := "Overview"
		if err := f.SetSheetName("Sheet1", sheet); err != nil {
			log.Printf("taxes.Summary SetSheetName: %s", err)
		}

		setCellValue(f, sheet, "A1", "Revenue (incl. tax)")
		setCellValue(f, sheet, "B1", "Revenue (excl. tax)")
		setCellValue(f, sheet, "C1", "Tax")
		setCellValue(f, sheet, "D1", "Hours")

		setCellFloat(f, sheet, "A2", export.TotalRevenue)
		setCellFloat(f, sheet, "B2", export.TotalEx)
		setCellFloat(f, sheet, "C2", export.TotalTax)
		setCellFloat(f, sheet, "D2", export.TotalHours)

		if quarter > 0 {
			setCellValue(f, sheet, "A4", fmt.Sprintf("Filtered: Q%d only", quarter))
		}
	}

	// Sheet 2: Companies (sorted, with Type column)
	{
		sheet := "Companies"
		newSheet(f, sheet)

		setCellValue(f, sheet, "A1", "Company")
		setCellValue(f, sheet, "B1", "VAT Number")
		setCellValue(f, sheet, "C1", "Type")
		setCellValue(f, sheet, "D1", "Revenue")
		setCellValue(f, sheet, "E1", "AccountingCode")

		// Sort companies by name
		sort.Slice(export.Companies, func(i, j int) bool {
			return export.Companies[i].Name < export.Companies[j].Name
		})

		var totalRevenue float64
		pos := 1
		for _, c := range export.Companies {
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), c.Name)
			setCellValue(f, sheet, fmt.Sprintf("B%d", pos), c.VAT)
			setCellValue(f, sheet, fmt.Sprintf("C%d", pos), c.TaxCategory)
			setCellFloat(f, sheet, fmt.Sprintf("D%d", pos), c.TotalRevenue)
			setCellValue(f, sheet, fmt.Sprintf("E%d", pos), c.AccountingCode)
			totalRevenue += c.TotalRevenue
		}

		// Totals row
		pos++
		setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "TOTAL")
		setCellFloat(f, sheet, fmt.Sprintf("D%d", pos), totalRevenue)
	}

	// Sheet 3: Invoices (sorted by date, with Status and Type)
	{
		sheet := "Invoices"
		newSheet(f, sheet)

		setCellValue(f, sheet, "A1", "Date")
		setCellValue(f, sheet, "B1", "InvoiceID")
		setCellValue(f, sheet, "C1", "AccountingID")
		setCellValue(f, sheet, "D1", "Customer")
		setCellValue(f, sheet, "E1", "Type")
		setCellValue(f, sheet, "F1", "Status")
		setCellValue(f, sheet, "G1", "Ex")
		setCellValue(f, sheet, "H1", "Tax")
		setCellValue(f, sheet, "I1", "Total")
		setCellValue(f, sheet, "J1", "AccountingCode")

		var totalEx, totalTax, totalInc float64
		pos := 1
		for _, inv := range export.Invoices {
			pos++
			// Extract accounting ID from invoice ID (e.g., "2024Q1-0152" -> "20240152")
			acctID := fmt.Sprintf("%d%s", year, extractInvoiceNum(inv.InvoiceID))

			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), inv.Issuedate)
			setCellValue(f, sheet, fmt.Sprintf("B%d", pos), inv.InvoiceID)
			setCellValue(f, sheet, fmt.Sprintf("C%d", pos), acctID)
			setCellValue(f, sheet, fmt.Sprintf("D%d", pos), inv.CustomerName)
			setCellValue(f, sheet, fmt.Sprintf("E%d", pos), inv.TaxCategory)
			setCellValue(f, sheet, fmt.Sprintf("F%d", pos), inv.Status)
			setCellFloat(f, sheet, fmt.Sprintf("G%d", pos), inv.TotalEx)
			setCellFloat(f, sheet, fmt.Sprintf("H%d", pos), inv.TotalTax)
			setCellFloat(f, sheet, fmt.Sprintf("I%d", pos), inv.TotalInc)
			setCellValue(f, sheet, fmt.Sprintf("J%d", pos), inv.AccountingCode)

			totalEx += inv.TotalEx
			totalTax += inv.TotalTax
			totalInc += inv.TotalInc
		}

		// Totals row
		pos++
		setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "TOTAL")
		setCellFloat(f, sheet, fmt.Sprintf("G%d", pos), totalEx)
		setCellFloat(f, sheet, fmt.Sprintf("H%d", pos), totalTax)
		setCellFloat(f, sheet, fmt.Sprintf("I%d", pos), totalInc)
	}

	// Sheet 4: EU Companies (ICP report - only EU0)
	{
		sheet := "ICP-EU"
		newSheet(f, sheet)

		setCellValue(f, sheet, "A1", "Company")
		setCellValue(f, sheet, "B1", "VAT Number")
		setCellValue(f, sheet, "C1", "Revenue")

		var totalEU float64
		pos := 1
		for _, c := range export.Companies {
			if c.TaxCategory != "EU0" {
				continue
			}
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), c.Name)
			setCellValue(f, sheet, fmt.Sprintf("B%d", pos), c.VAT)
			setCellFloat(f, sheet, fmt.Sprintf("C%d", pos), c.TotalRevenue)
			totalEU += c.TotalRevenue
		}

		// Totals row
		if pos > 1 {
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "TOTAL")
			setCellFloat(f, sheet, fmt.Sprintf("C%d", pos), totalEU)
		}
	}

	// Sheet 5: AccountingSales (for import into accounting software)
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
		for _, inv := range export.Invoices {
			acctID := fmt.Sprintf("%d%s", year, extractInvoiceNum(inv.InvoiceID))

			// Determine ledgers based on tax category
			// 1300 = Debtors (debet = total incl)
			// 1671 = VAT payable (credit = tax) - only for NL invoices
			// 8000 = Revenue (credit = ex)
			ledgers := []string{"1300", "8000"}
			if inv.TaxCategory == "NL" && inv.TotalTax > 0 {
				ledgers = []string{"1300", "1671", "8000"}
			}

			for _, ledger := range ledgers {
				pos++

				setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "1300")
				setCellValue(f, sheet, fmt.Sprintf("B%d", pos), acctID)
				setCellValue(f, sheet, fmt.Sprintf("C%d", pos), inv.Issuedate)
				setCellValue(f, sheet, fmt.Sprintf("D%d", pos), ledger)

				// Debet: only for ledger 1300 (receivables)
				if ledger == "1300" {
					setCellFloat(f, sheet, fmt.Sprintf("E%d", pos), inv.TotalInc)
				} else {
					setCellValue(f, sheet, fmt.Sprintf("E%d", pos), "")
				}

				// Credit: tax for 1671, ex for 8000
				switch ledger {
				case "1671":
					setCellFloat(f, sheet, fmt.Sprintf("F%d", pos), inv.TotalTax)
				case "8000":
					setCellFloat(f, sheet, fmt.Sprintf("F%d", pos), inv.TotalEx)
				default:
					setCellValue(f, sheet, fmt.Sprintf("F%d", pos), "")
				}

				setCellValue(f, sheet, fmt.Sprintf("G%d", pos), "")
				setCellValue(f, sheet, fmt.Sprintf("H%d", pos), "")
				setCellValue(f, sheet, fmt.Sprintf("I%d", pos), acctID)
				setCellValue(f, sheet, fmt.Sprintf("J%d", pos), inv.CustomerName)

				// Relation code from debtor
				acctCode := inv.AccountingCode
				if acctCode == "" {
					acctCode = "0"
					log.Printf("WARN: Missing AccountingCode for customer %s", inv.CustomerName)
				}
				setCellValue(f, sheet, fmt.Sprintf("K%d", pos), acctCode)

				setCellValue(f, sheet, fmt.Sprintf("L%d", pos), inv.InvoiceID)
				setCellValue(f, sheet, fmt.Sprintf("M%d", pos), "")
			}
		}
	}

	// Generate filename
	fname := fmt.Sprintf("%s-%d", entity, year)
	if quarter > 0 {
		fname = fmt.Sprintf("%s-%d-Q%d", entity, year, quarter)
	}
	fname += ".xlsx"

	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Header().Set("Content-Disposition", "attachment; filename="+fname)
	if _, e := f.WriteTo(w); e != nil {
		httputil.LogErr("taxes.Summary excel.WriteTo", e)
	}
}

// extractInvoiceNum extracts the numeric part from an invoice ID (e.g., "2024Q1-0152" -> "0152")
func extractInvoiceNum(invoiceID string) string {
	parts := strings.Split(invoiceID, "-")
	if len(parts) >= 2 {
		return parts[1]
	}
	return invoiceID
}
