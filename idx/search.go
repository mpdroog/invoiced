package idx

import (
	"context"
	"log"
	"strconv"
	"strings"
)

// SearchResult represents a single search result item.
type SearchResult struct {
	Type     string `json:"type"`     // "invoice", "hour", "purchase"
	ID       string `json:"id"`       // conceptid/filename
	Title    string `json:"title"`    // Display title
	Subtitle string `json:"subtitle"` // Additional info
	Entity   string `json:"entity"`
	Year     int    `json:"year"`
	Quarter  int    `json:"quarter"`
	Bucket   string `json:"bucket"` // For building URLs
}

// Search searches across invoices, hours, and purchase invoices
func Search(entity string, query string, limit int) ([]SearchResult, error) {
	if DB == nil {
		return nil, nil
	}

	if limit <= 0 {
		limit = 20
	}

	query = "%" + strings.ToLower(query) + "%"
	var results []SearchResult

	// Search invoices
	rows, err := DB.QueryContext(context.Background(), `
		SELECT id, entity, year, quarter, status, customer_name, invoiceid, total_inc
		FROM invoices
		WHERE entity = ? AND (
			LOWER(customer_name) LIKE ? OR
			LOWER(invoiceid) LIKE ? OR
			LOWER(id) LIKE ? OR
			LOWER(notes) LIKE ?
		)
		ORDER BY year DESC, quarter DESC
		LIMIT ?`,
		entity, query, query, query, query, limit,
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
		var r SearchResult
		var customerName, invoiceID, totalInc, status string
		if err := rows.Scan(&r.ID, &r.Entity, &r.Year, &r.Quarter, &status, &customerName, &invoiceID, &totalInc); err != nil {
			return nil, err
		}
		r.Type = "invoice"
		if invoiceID != "" {
			r.Title = invoiceID + " - " + customerName
		} else {
			r.Title = r.ID + " - " + customerName
		}
		r.Subtitle = "€ " + totalInc

		// Determine bucket for URL
		switch status {
		case "CONCEPT":
			r.Bucket = bucketConcept
		case "PAID":
			r.Bucket = "Q" + strconv.Itoa(r.Quarter)
		case "UNPAID":
			r.Bucket = "Q" + strconv.Itoa(r.Quarter)
		default:
			r.Bucket = bucketConcept
		}

		results = append(results, r)
	}

	if len(results) >= limit {
		return results, nil
	}

	// Search hours
	rows2, err := DB.QueryContext(context.Background(), `
		SELECT id, entity, year, quarter, status, project, name, total_hours
		FROM hours
		WHERE entity = ? AND (
			LOWER(project) LIKE ? OR
			LOWER(name) LIKE ? OR
			LOWER(id) LIKE ?
		)
		ORDER BY year DESC, quarter DESC
		LIMIT ?`,
		entity, query, query, query, limit-len(results),
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
		var r SearchResult
		var project, name, totalHours, status string
		if err := rows2.Scan(&r.ID, &r.Entity, &r.Year, &r.Quarter, &status, &project, &name, &totalHours); err != nil {
			return nil, err
		}
		r.Type = "hour"
		r.Title = project
		if name != "" {
			r.Title += " - " + name
		}
		r.Subtitle = totalHours + " hours"

		if status == "CONCEPT" || r.Quarter == 0 {
			r.Bucket = bucketConcept
		} else {
			r.Bucket = "Q" + strconv.Itoa(r.Quarter)
		}

		results = append(results, r)
	}

	if len(results) >= limit {
		return results, nil
	}

	// Search purchase invoices
	rows3, err := DB.QueryContext(context.Background(), `
		SELECT id, entity, year, quarter, status, supplier_name, invoiceid, total_inc
		FROM purchase_invoices
		WHERE entity = ? AND (
			LOWER(supplier_name) LIKE ? OR
			LOWER(invoiceid) LIKE ? OR
			LOWER(id) LIKE ?
		)
		ORDER BY year DESC, quarter DESC
		LIMIT ?`,
		entity, query, query, query, limit-len(results),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows3.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	for rows3.Next() {
		var r SearchResult
		var supplierName, invoiceID, totalInc, status string
		if err := rows3.Scan(&r.ID, &r.Entity, &r.Year, &r.Quarter, &status, &supplierName, &invoiceID, &totalInc); err != nil {
			return nil, err
		}
		r.Type = "purchase"
		r.Title = invoiceID + " - " + supplierName
		r.Subtitle = "€ " + totalInc

		r.Bucket = "Q" + strconv.Itoa(r.Quarter)

		results = append(results, r)
	}

	return results, nil
}
