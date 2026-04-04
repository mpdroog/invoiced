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

	// Create styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E0E0E0"}, Pattern: 1},
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFACD"}, Pattern: 1},
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4, // #,##0.00
	})

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
		setRowStyle(f, sheet, 1, headerStyle)

		setCellFloat(f, sheet, "A2", export.TotalRevenue)
		setCellFloat(f, sheet, "B2", export.TotalEx)
		setCellFloat(f, sheet, "C2", export.TotalTax)
		setCellFloat(f, sheet, "D2", export.TotalHours)

		if quarter > 0 {
			setCellValue(f, sheet, "A4", fmt.Sprintf("Filtered: Q%d only", quarter))
		}

		setColWidths(f, sheet, map[string]float64{"A": 20, "B": 20, "C": 15, "D": 10})
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
		setRowStyle(f, sheet, 1, headerStyle)
		freezeRow(f, sheet)

		// Sort companies by name
		sort.Slice(export.Companies, func(i, j int) bool {
			return export.Companies[i].Name < export.Companies[j].Name
		})

		pos := 1
		for _, c := range export.Companies {
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), c.Name)
			setCellValue(f, sheet, fmt.Sprintf("B%d", pos), c.VAT)
			setCellValue(f, sheet, fmt.Sprintf("C%d", pos), c.TaxCategory)
			setCellFloat(f, sheet, fmt.Sprintf("D%d", pos), c.TotalRevenue)
			setCellValue(f, sheet, fmt.Sprintf("E%d", pos), c.AccountingCode)
		}

		// Totals row with SUM formula
		pos++
		setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "TOTAL")
		setCellFormula(f, sheet, fmt.Sprintf("D%d", pos), fmt.Sprintf("SUM(D2:D%d)", pos-1))
		setRowStyle(f, sheet, pos, totalStyle)
		setColStyle(f, sheet, "D", moneyStyle)

		setColWidths(f, sheet, map[string]float64{"A": 30, "B": 20, "C": 10, "D": 15, "E": 15})
	}

	// Sheet 3: Invoices (sorted by date, with Status, Type, and Paydate)
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
		setCellValue(f, sheet, "J1", "AcctCode")
		setCellValue(f, sheet, "K1", "PayDate")
		setRowStyle(f, sheet, 1, headerStyle)
		freezeRow(f, sheet)

		pos := 1
		for _, inv := range export.Invoices {
			pos++
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
			setCellValue(f, sheet, fmt.Sprintf("K%d", pos), inv.Paydate)
		}

		// Totals row with SUM formulas
		pos++
		setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "TOTAL")
		setCellFormula(f, sheet, fmt.Sprintf("G%d", pos), fmt.Sprintf("SUM(G2:G%d)", pos-1))
		setCellFormula(f, sheet, fmt.Sprintf("H%d", pos), fmt.Sprintf("SUM(H2:H%d)", pos-1))
		setCellFormula(f, sheet, fmt.Sprintf("I%d", pos), fmt.Sprintf("SUM(I2:I%d)", pos-1))
		setRowStyle(f, sheet, pos, totalStyle)

		for _, col := range []string{"G", "H", "I"} {
			setColStyle(f, sheet, col, moneyStyle)
		}
		setColWidths(f, sheet, map[string]float64{
			"A": 12, "B": 15, "C": 12, "D": 25, "E": 8,
			"F": 8, "G": 12, "H": 12, "I": 12, "J": 10, "K": 12,
		})
	}

	// Sheet 4: EU Companies (ICP report - only EU0)
	{
		sheet := "ICP-EU"
		newSheet(f, sheet)

		setCellValue(f, sheet, "A1", "Company")
		setCellValue(f, sheet, "B1", "VAT Number")
		setCellValue(f, sheet, "C1", "Revenue")
		setRowStyle(f, sheet, 1, headerStyle)
		freezeRow(f, sheet)

		pos := 1
		for _, c := range export.Companies {
			if c.TaxCategory != "EU0" {
				continue
			}
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), c.Name)
			setCellValue(f, sheet, fmt.Sprintf("B%d", pos), c.VAT)
			setCellFloat(f, sheet, fmt.Sprintf("C%d", pos), c.TotalRevenue)
		}

		// Totals row with SUM formula
		if pos > 1 {
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "TOTAL")
			setCellFormula(f, sheet, fmt.Sprintf("C%d", pos), fmt.Sprintf("SUM(C2:C%d)", pos-1))
			setRowStyle(f, sheet, pos, totalStyle)
		}

		setColStyle(f, sheet, "C", moneyStyle)
		setColWidths(f, sheet, map[string]float64{"A": 30, "B": 20, "C": 15})
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
		setRowStyle(f, sheet, 1, headerStyle)
		freezeRow(f, sheet)

		pos := 1
		for _, inv := range export.Invoices {
			acctID := fmt.Sprintf("%d%s", year, extractInvoiceNum(inv.InvoiceID))

			// Determine ledgers based on tax category
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

				if ledger == "1300" {
					setCellFloat(f, sheet, fmt.Sprintf("E%d", pos), inv.TotalInc)
				} else {
					setCellValue(f, sheet, fmt.Sprintf("E%d", pos), "")
				}

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

		for _, col := range []string{"E", "F"} {
			setColStyle(f, sheet, col, moneyStyle)
		}
	}

	// Sheet 6: AccountingPurchases (Voorbelasting / purchase invoices)
	{
		sheet := "AccountingPurchases"
		newSheet(f, sheet)

		setCellValue(f, sheet, "A1", "Date")
		setCellValue(f, sheet, "B1", "InvoiceID")
		setCellValue(f, sheet, "C1", "Supplier")
		setCellValue(f, sheet, "D1", "VAT Number")
		setCellValue(f, sheet, "E1", "Status")
		setCellValue(f, sheet, "F1", "Ex")
		setCellValue(f, sheet, "G1", "Tax")
		setCellValue(f, sheet, "H1", "Total")
		setCellValue(f, sheet, "I1", "PayDate")
		setCellValue(f, sheet, "J1", "PaymentRef")
		setCellValue(f, sheet, "K1", "IBAN")
		setRowStyle(f, sheet, 1, headerStyle)
		freezeRow(f, sheet)

		pos := 1
		for _, p := range export.Purchases {
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), p.Issuedate)
			setCellValue(f, sheet, fmt.Sprintf("B%d", pos), p.InvoiceID)
			setCellValue(f, sheet, fmt.Sprintf("C%d", pos), p.SupplierName)
			setCellValue(f, sheet, fmt.Sprintf("D%d", pos), p.SupplierVAT)
			setCellValue(f, sheet, fmt.Sprintf("E%d", pos), p.Status)
			setCellFloat(f, sheet, fmt.Sprintf("F%d", pos), p.TotalEx)
			setCellFloat(f, sheet, fmt.Sprintf("G%d", pos), p.TotalTax)
			setCellFloat(f, sheet, fmt.Sprintf("H%d", pos), p.TotalInc)
			setCellValue(f, sheet, fmt.Sprintf("I%d", pos), p.Paydate)
			setCellValue(f, sheet, fmt.Sprintf("J%d", pos), p.PaymentRef)
			setCellValue(f, sheet, fmt.Sprintf("K%d", pos), p.IBAN)
		}

		// Totals row with SUM formulas
		if pos > 1 {
			pos++
			setCellValue(f, sheet, fmt.Sprintf("A%d", pos), "TOTAL")
			setCellFormula(f, sheet, fmt.Sprintf("F%d", pos), fmt.Sprintf("SUM(F2:F%d)", pos-1))
			setCellFormula(f, sheet, fmt.Sprintf("G%d", pos), fmt.Sprintf("SUM(G2:G%d)", pos-1))
			setCellFormula(f, sheet, fmt.Sprintf("H%d", pos), fmt.Sprintf("SUM(H2:H%d)", pos-1))
			setRowStyle(f, sheet, pos, totalStyle)
		}

		for _, col := range []string{"F", "G", "H"} {
			setColStyle(f, sheet, col, moneyStyle)
		}
		setColWidths(f, sheet, map[string]float64{
			"A": 12, "B": 20, "C": 25, "D": 18, "E": 8,
			"F": 12, "G": 12, "H": 12, "I": 12, "J": 20, "K": 25,
		})
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

// setCellFormula sets a cell formula
func setCellFormula(f *excelize.File, sheet, cell, formula string) {
	if err := f.SetCellFormula(sheet, cell, formula); err != nil {
		log.Printf("taxes.Summary SetCellFormula %s:%s: %s", sheet, cell, err)
	}
}

// setRowStyle applies a style to an entire row
func setRowStyle(f *excelize.File, sheet string, row, styleID int) {
	if err := f.SetRowStyle(sheet, row, row, styleID); err != nil {
		log.Printf("taxes.Summary SetRowStyle %s:%d: %s", sheet, row, err)
	}
}

// setColStyle applies a style to an entire column
func setColStyle(f *excelize.File, sheet, col string, styleID int) {
	if err := f.SetColStyle(sheet, col, styleID); err != nil {
		log.Printf("taxes.Summary SetColStyle %s:%s: %s", sheet, col, err)
	}
}

// setColWidths sets column widths from a map
func setColWidths(f *excelize.File, sheet string, widths map[string]float64) {
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			log.Printf("taxes.Summary SetColWidth %s:%s: %s", sheet, col, err)
		}
	}
}

// freezeRow freezes the first row (header) for scrolling
func freezeRow(f *excelize.File, sheet string) {
	if err := f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		log.Printf("taxes.Summary SetPanes %s: %s", sheet, err)
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
