package idx

import (
	"context"
	"database/sql"
	"log"

	"github.com/mpdroog/invoiced/model"
	"github.com/shopspring/decimal"
)

const zeroDecimal = "0.00"

// ctx returns a background context for database operations
func ctx() context.Context {
	return context.Background()
}

// GetQuarterTaxSummary returns aggregated tax data for a specific quarter
func GetQuarterTaxSummary(entity string, year int, quarter int) (*model.TaxSummary, []model.InvoiceAuditLine, error) {
	if DB == nil {
		return nil, nil, nil
	}

	sum := &model.TaxSummary{
		EUCompany: make(map[string]string),
	}
	var audit []model.InvoiceAuditLine

	// Initialize decimals
	nlEx := decimal.Zero
	nlTax := decimal.Zero
	euEx := decimal.Zero
	worldEx := decimal.Zero
	totalRevenue := decimal.Zero

	// Query all invoices for this quarter (both paid and unpaid)
	rows, err := DB.QueryContext(ctx(), `
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

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
		audit = append(audit, model.InvoiceAuditLine{
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

// GetYearlyCustomerTotals returns revenue totals grouped by customer for a year
func GetYearlyCustomerTotals(entity string, year int, paidOnly bool) ([]model.CustomerTotal, error) {
	if DB == nil {
		return nil, nil
	}

	statusFilter := "status IN ('PAID', 'UNPAID')"
	if paidOnly {
		statusFilter = "status = 'PAID'"
	}

	rows, err := DB.QueryContext(ctx(), `
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var results []model.CustomerTotal
	for rows.Next() {
		var ct model.CustomerTotal
		var revenue float64
		if err := rows.Scan(&ct.Name, &revenue, &ct.InvoiceCount); err != nil {
			return nil, err
		}
		ct.Revenue = decimal.NewFromFloat(revenue).StringFixed(2)
		results = append(results, ct)
	}

	return results, rows.Err()
}

// GetYearlyQuarterSummary returns summary data for all quarters in a year
func GetYearlyQuarterSummary(entity string, year int) ([]model.QuarterSummary, error) {
	if DB == nil {
		return nil, nil
	}

	rows, err := DB.QueryContext(ctx(), `
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var results []model.QuarterSummary
	for rows.Next() {
		var qs model.QuarterSummary
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

// GetMonthlyMetrics returns revenue and hours grouped by month for a year
func GetMonthlyMetrics(entity string, year int) (map[string]*model.MonthlyMetric, error) {
	if DB == nil {
		return nil, nil
	}

	m := make(map[string]*model.MonthlyMetric)

	// Get revenue per month from paid invoices (using issuedate)
	rows, err := DB.QueryContext(ctx(), `
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	for rows.Next() {
		var yearmonth string
		var total, ex float64
		if err := rows.Scan(&yearmonth, &total, &ex); err != nil {
			return nil, err
		}
		m[yearmonth] = &model.MonthlyMetric{
			RevenueTotal: decimal.NewFromFloat(total).StringFixed(2),
			RevenueEx:    decimal.NewFromFloat(ex).StringFixed(2),
			Hours:        zeroDecimal,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get hours per month
	rows2, err := DB.QueryContext(ctx(), `
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
	defer func() {
		if err := rows2.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	for rows2.Next() {
		var yearmonth string
		var hours float64
		if err := rows2.Scan(&yearmonth, &hours); err != nil {
			return nil, err
		}
		if _, ok := m[yearmonth]; !ok {
			m[yearmonth] = &model.MonthlyMetric{
				RevenueTotal: zeroDecimal,
				RevenueEx:    zeroDecimal,
			}
		}
		m[yearmonth].Hours = decimal.NewFromFloat(hours).StringFixed(2)
	}

	return m, rows2.Err()
}

// GetYearlyTotal returns total revenue for an entity/year
func GetYearlyTotal(entity string, year int) (string, error) {
	if DB == nil {
		return zeroDecimal, nil
	}

	var total sql.NullFloat64
	err := DB.QueryRowContext(ctx(), `
		SELECT SUM(CAST(total_ex AS REAL))
		FROM invoices
		WHERE entity = ? AND year = ? AND status IN ('PAID', 'UNPAID')`,
		entity, year,
	).Scan(&total)

	if err != nil {
		return zeroDecimal, err
	}

	if !total.Valid {
		return zeroDecimal, nil
	}

	return decimal.NewFromFloat(total.Float64).StringFixed(2), nil
}

// GetUnpaidSummary returns count and total of unpaid invoices
func GetUnpaidSummary(entity string, year int) (*model.UnpaidSummary, error) {
	if DB == nil {
		return &model.UnpaidSummary{Count: 0, TotalAmount: zeroDecimal}, nil
	}

	var count int
	var total sql.NullFloat64
	err := DB.QueryRowContext(ctx(), `
		SELECT COUNT(*), COALESCE(SUM(CAST(total_inc AS REAL)), 0)
		FROM invoices
		WHERE entity = ? AND year = ? AND status = 'UNPAID'`,
		entity, year,
	).Scan(&count, &total)

	if err != nil {
		return nil, err
	}

	return &model.UnpaidSummary{
		Count:       count,
		TotalAmount: decimal.NewFromFloat(total.Float64).StringFixed(2),
	}, nil
}

// GetOverdueInvoices returns invoices past their due date
func GetOverdueInvoices(entity string, year int, today string) ([]model.OverdueInvoice, error) {
	if DB == nil {
		return nil, nil
	}

	rows, err := DB.QueryContext(ctx(), `
		SELECT id, invoiceid, customer_name, duedate, total_inc, quarter,
		       julianday(?) - julianday(duedate) as days_overdue
		FROM invoices
		WHERE entity = ? AND year = ? AND status = 'UNPAID'
		  AND duedate != '' AND duedate < ?
		ORDER BY duedate ASC`,
		today, entity, year, today,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var results []model.OverdueInvoice
	for rows.Next() {
		var inv model.OverdueInvoice
		var daysOverdue float64
		if err := rows.Scan(&inv.ID, &inv.InvoiceID, &inv.CustomerName, &inv.DueDate, &inv.Amount, &inv.Quarter, &daysOverdue); err != nil {
			return nil, err
		}
		inv.DaysOverdue = int(daysOverdue)
		results = append(results, inv)
	}

	return results, rows.Err()
}

// GetUnbilledHours returns count and total of concept hours (not yet billed)
func GetUnbilledHours(entity string, year int) (*model.UnbilledHoursSummary, error) {
	if DB == nil {
		return &model.UnbilledHoursSummary{Count: 0, TotalHours: zeroDecimal}, nil
	}

	var count int
	var total sql.NullFloat64
	err := DB.QueryRowContext(ctx(), `
		SELECT COUNT(*), COALESCE(SUM(CAST(total_hours AS REAL)), 0)
		FROM hours
		WHERE entity = ? AND year = ? AND status = 'CONCEPT'`,
		entity, year,
	).Scan(&count, &total)

	if err != nil {
		return nil, err
	}

	return &model.UnbilledHoursSummary{
		Count:      count,
		TotalHours: decimal.NewFromFloat(total.Float64).StringFixed(2),
	}, nil
}

// GetYearComparison compares revenue between current and previous year
func GetYearComparison(entity string, currentYear int) (*model.YearComparison, error) {
	if DB == nil {
		return nil, nil
	}

	previousYear := currentYear - 1

	var currentTotal, previousTotal sql.NullFloat64

	// Current year revenue
	err := DB.QueryRowContext(ctx(), `
		SELECT COALESCE(SUM(CAST(total_ex AS REAL)), 0)
		FROM invoices
		WHERE entity = ? AND year = ? AND status IN ('PAID', 'UNPAID')`,
		entity, currentYear,
	).Scan(&currentTotal)
	if err != nil {
		return nil, err
	}

	// Previous year revenue
	err = DB.QueryRowContext(ctx(), `
		SELECT COALESCE(SUM(CAST(total_ex AS REAL)), 0)
		FROM invoices
		WHERE entity = ? AND year = ? AND status IN ('PAID', 'UNPAID')`,
		entity, previousYear,
	).Scan(&previousTotal)
	if err != nil {
		return nil, err
	}

	current := decimal.NewFromFloat(currentTotal.Float64)
	previous := decimal.NewFromFloat(previousTotal.Float64)
	growth := current.Sub(previous)

	growthPct := decimal.Zero
	if !previous.IsZero() {
		growthPct = growth.Div(previous).Mul(decimal.NewFromInt(100))
	}

	return &model.YearComparison{
		CurrentYear:     currentYear,
		PreviousYear:    previousYear,
		CurrentRevenue:  current.StringFixed(2),
		PreviousRevenue: previous.StringFixed(2),
		GrowthPercent:   growthPct.StringFixed(1),
		GrowthAmount:    growth.StringFixed(2),
	}, nil
}
