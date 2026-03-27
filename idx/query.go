package idx

import (
	"database/sql"

	"github.com/shopspring/decimal"
)

// TaxSummary contains aggregated tax data for a quarter
type TaxSummary struct {
	Ex        string            // Sum revenue of NL invoices
	Tax       string            // Tax to pay
	EUEx      string            // Sum revenue of EU invoices
	EUCompany map[string]string // Revenue per EU company VAT number (for ICP)
	ExWorld   string            // Sum revenue of world invoices
	ExRevenue string            // Sum revenue of everything
}

// InvoiceAuditLine contains info for audit logging
type InvoiceAuditLine struct {
	InvoiceID   string
	TaxCategory string
	TotalEx     string
	TotalTax    string
}

// GetQuarterTaxSummary returns aggregated tax data for a specific quarter
func GetQuarterTaxSummary(entity string, year int, quarter int) (*TaxSummary, []InvoiceAuditLine, error) {
	if DB == nil {
		return nil, nil, nil
	}

	sum := &TaxSummary{
		EUCompany: make(map[string]string),
	}
	var audit []InvoiceAuditLine

	// Initialize decimals
	nlEx := decimal.Zero
	nlTax := decimal.Zero
	euEx := decimal.Zero
	worldEx := decimal.Zero
	totalRevenue := decimal.Zero

	// Query all invoices for this quarter (both paid and unpaid)
	rows, err := DB.Query(`
		SELECT id, invoiceid, tax_category, total_ex, total_tax, total_inc, customer_vat
		FROM invoices
		WHERE entity = ? AND year = ? AND quarter = ?
		  AND status IN ('UNPAID', 'PAID')
		ORDER BY invoiceid`,
		entity, year, quarter,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, invoiceid, taxCat, totalEx, totalTax, totalInc string
		var customerVat sql.NullString

		if err := rows.Scan(&id, &invoiceid, &taxCat, &totalEx, &totalTax, &totalInc, &customerVat); err != nil {
			return nil, nil, err
		}

		ex, err := decimal.NewFromString(totalEx)
		if err != nil {
			ex = decimal.Zero
		}
		tax, err := decimal.NewFromString(totalTax)
		if err != nil {
			tax = decimal.Zero
		}
		inc, err := decimal.NewFromString(totalInc)
		if err != nil {
			inc = decimal.Zero
		}

		// Add to audit log
		audit = append(audit, InvoiceAuditLine{
			InvoiceID:   invoiceid,
			TaxCategory: taxCat,
			TotalEx:     totalEx,
			TotalTax:    totalTax,
		})

		// Aggregate by category
		switch taxCat {
		case "WORLD0":
			worldEx = worldEx.Add(ex)
		case "EU0":
			euEx = euEx.Add(ex)
			// Track per EU company for ICP
			if customerVat.Valid && customerVat.String != "" {
				existing := decimal.Zero
				if v, ok := sum.EUCompany[customerVat.String]; ok {
					existing, _ = decimal.NewFromString(v)
				}
				sum.EUCompany[customerVat.String] = existing.Add(inc).StringFixed(2)
			}
		default: // NL
			nlEx = nlEx.Add(ex)
			nlTax = nlTax.Add(tax)
		}

		totalRevenue = totalRevenue.Add(ex)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Format results
	sum.Ex = nlEx.StringFixed(2)
	sum.Tax = nlTax.StringFixed(2)
	sum.EUEx = euEx.StringFixed(2)
	sum.ExWorld = worldEx.StringFixed(2)
	sum.ExRevenue = totalRevenue.StringFixed(2)

	return sum, audit, nil
}

// CustomerTotal represents revenue per customer
type CustomerTotal struct {
	Name         string
	Revenue      string
	InvoiceCount int
}

// GetYearlyCustomerTotals returns revenue totals grouped by customer for a year
func GetYearlyCustomerTotals(entity string, year int, paidOnly bool) ([]CustomerTotal, error) {
	if DB == nil {
		return nil, nil
	}

	statusFilter := "status IN ('PAID', 'UNPAID')"
	if paidOnly {
		statusFilter = "status = 'PAID'"
	}

	rows, err := DB.Query(`
		SELECT customer_name, SUM(CAST(total_ex AS REAL)) as revenue, COUNT(*) as cnt
		FROM invoices
		WHERE entity = ? AND year = ? AND `+statusFilter+`
		GROUP BY customer_name
		ORDER BY revenue DESC`,
		entity, year,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CustomerTotal
	for rows.Next() {
		var ct CustomerTotal
		var revenue float64
		if err := rows.Scan(&ct.Name, &revenue, &ct.InvoiceCount); err != nil {
			return nil, err
		}
		ct.Revenue = decimal.NewFromFloat(revenue).StringFixed(2)
		results = append(results, ct)
	}

	return results, rows.Err()
}

// QuarterSummary contains aggregated data for a quarter
type QuarterSummary struct {
	Quarter       int
	InvoiceCount  int
	TotalRevenue  string
	TotalTax      string
	PaidCount     int
	UnpaidCount   int
	PaidRevenue   string
	UnpaidRevenue string
}

// GetYearlyQuarterSummary returns summary data for all quarters in a year
func GetYearlyQuarterSummary(entity string, year int) ([]QuarterSummary, error) {
	if DB == nil {
		return nil, nil
	}

	rows, err := DB.Query(`
		SELECT
			quarter,
			COUNT(*) as cnt,
			COALESCE(SUM(CAST(total_ex AS REAL)), 0) as revenue,
			COALESCE(SUM(CAST(total_tax AS REAL)), 0) as tax,
			SUM(CASE WHEN status = 'PAID' THEN 1 ELSE 0 END) as paid_cnt,
			SUM(CASE WHEN status = 'UNPAID' THEN 1 ELSE 0 END) as unpaid_cnt,
			COALESCE(SUM(CASE WHEN status = 'PAID' THEN CAST(total_ex AS REAL) ELSE 0 END), 0) as paid_rev,
			COALESCE(SUM(CASE WHEN status = 'UNPAID' THEN CAST(total_ex AS REAL) ELSE 0 END), 0) as unpaid_rev
		FROM invoices
		WHERE entity = ? AND year = ? AND quarter > 0
		GROUP BY quarter
		ORDER BY quarter`,
		entity, year,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []QuarterSummary
	for rows.Next() {
		var qs QuarterSummary
		var revenue, tax, paidRev, unpaidRev float64
		if err := rows.Scan(&qs.Quarter, &qs.InvoiceCount, &revenue, &tax,
			&qs.PaidCount, &qs.UnpaidCount, &paidRev, &unpaidRev); err != nil {
			return nil, err
		}
		qs.TotalRevenue = decimal.NewFromFloat(revenue).StringFixed(2)
		qs.TotalTax = decimal.NewFromFloat(tax).StringFixed(2)
		qs.PaidRevenue = decimal.NewFromFloat(paidRev).StringFixed(2)
		qs.UnpaidRevenue = decimal.NewFromFloat(unpaidRev).StringFixed(2)
		results = append(results, qs)
	}

	return results, rows.Err()
}

// MonthlyMetric contains revenue and hours for a month
type MonthlyMetric struct {
	RevenueTotal string
	RevenueEx    string
	Hours        string
}

// GetMonthlyMetrics returns revenue and hours grouped by month for a year
func GetMonthlyMetrics(entity string, year int) (map[string]*MonthlyMetric, error) {
	if DB == nil {
		return nil, nil
	}

	m := make(map[string]*MonthlyMetric)

	// Get revenue per month from paid invoices (using issuedate)
	rows, err := DB.Query(`
		SELECT substr(issuedate, 1, 7) as yearmonth,
		       COALESCE(SUM(CAST(total_inc AS REAL)), 0) as total,
		       COALESCE(SUM(CAST(total_ex AS REAL)), 0) as ex
		FROM invoices
		WHERE entity = ? AND year = ? AND status = 'PAID' AND issuedate != ''
		GROUP BY yearmonth
		ORDER BY yearmonth`,
		entity, year,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var yearmonth string
		var total, ex float64
		if err := rows.Scan(&yearmonth, &total, &ex); err != nil {
			return nil, err
		}
		m[yearmonth] = &MonthlyMetric{
			RevenueTotal: decimal.NewFromFloat(total).StringFixed(2),
			RevenueEx:    decimal.NewFromFloat(ex).StringFixed(2),
			Hours:        "0.00",
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get hours per month
	rows2, err := DB.Query(`
		SELECT substr(issuedate, 1, 7) as yearmonth,
		       COALESCE(SUM(CAST(total_hours AS REAL)), 0) as hours
		FROM hours
		WHERE entity = ? AND year = ? AND issuedate != ''
		GROUP BY yearmonth
		ORDER BY yearmonth`,
		entity, year,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var yearmonth string
		var hours float64
		if err := rows2.Scan(&yearmonth, &hours); err != nil {
			return nil, err
		}
		if _, ok := m[yearmonth]; !ok {
			m[yearmonth] = &MonthlyMetric{
				RevenueTotal: "0.00",
				RevenueEx:    "0.00",
			}
		}
		m[yearmonth].Hours = decimal.NewFromFloat(hours).StringFixed(2)
	}

	return m, rows2.Err()
}

// GetYearlyTotal returns total revenue for an entity/year
func GetYearlyTotal(entity string, year int) (string, error) {
	if DB == nil {
		return "0.00", nil
	}

	var total sql.NullFloat64
	err := DB.QueryRow(`
		SELECT SUM(CAST(total_ex AS REAL))
		FROM invoices
		WHERE entity = ? AND year = ? AND status IN ('PAID', 'UNPAID')`,
		entity, year,
	).Scan(&total)

	if err != nil {
		return "0.00", err
	}

	if !total.Valid {
		return "0.00", nil
	}

	return decimal.NewFromFloat(total.Float64).StringFixed(2), nil
}
