package idx

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/model"
)

// PathParts contains parsed components from a TOML file path
type PathParts struct {
	FullPath string
	ID       string // filename without .toml
	Entity   string
	Year     int
	Quarter  int
	Status   string // CONCEPT, UNPAID, PAID
	Type     string // "invoice" or "hour"
}

// Patterns for matching invoice and hour paths
// e.g., "rootdev/2024/Q1/sales-invoices-paid/abc123.toml"
var (
	invoicePathRe = regexp.MustCompile(`^([^/]+)/(\d{4})/(Q[1-4]|concepts)/(sales-invoices[^/]*)/([^/]+)\.toml$`)
	hourPathRe    = regexp.MustCompile(`^([^/]+)/(\d{4})/(Q[1-4]|concepts)/hours/([^/]+)\.toml$`)
)

// parseInvoicePath extracts components from an invoice path
func parseInvoicePath(relPath string) *PathParts {
	matches := invoicePathRe.FindStringSubmatch(relPath)
	if matches == nil {
		return nil
	}

	year, _ := strconv.Atoi(matches[2])
	quarter := 0
	if matches[3] != "concepts" {
		q := matches[3][1] - '0' // Q1 -> 1
		quarter = int(q)
	}

	// Determine status from bucket name
	bucket := matches[4]
	status := "CONCEPT"
	if strings.Contains(bucket, "-paid") {
		status = "PAID"
	} else if strings.Contains(bucket, "-unpaid") {
		status = "UNPAID"
	}

	return &PathParts{
		ID:      matches[5],
		Entity:  matches[1],
		Year:    year,
		Quarter: quarter,
		Status:  status,
		Type:    "invoice",
	}
}

// parseHourPath extracts components from an hour path
func parseHourPath(relPath string) *PathParts {
	matches := hourPathRe.FindStringSubmatch(relPath)
	if matches == nil {
		return nil
	}

	year, _ := strconv.Atoi(matches[2])
	quarter := 0
	status := "CONCEPT"
	if matches[3] != "concepts" {
		q := matches[3][1] - '0'
		quarter = int(q)
		status = "FINAL"
	}

	return &PathParts{
		ID:      matches[4],
		Entity:  matches[1],
		Year:    year,
		Quarter: quarter,
		Status:  status,
		Type:    "hour",
	}
}

// SyncPath syncs a single file to the SQLite index
// relPath should be relative to the db root (e.g., "rootdev/2024/Q1/sales-invoices-paid/abc.toml")
func SyncPath(dbPath, relPath string) error {
	if DB == nil {
		return nil // Index not initialized
	}

	fullPath := filepath.Join(dbPath, relPath)

	// Try invoice first
	if parts := parseInvoicePath(relPath); parts != nil {
		parts.FullPath = fullPath
		return syncInvoice(parts)
	}

	// Try hour
	if parts := parseHourPath(relPath); parts != nil {
		parts.FullPath = fullPath
		return syncHour(parts)
	}

	// Not an indexed file type
	return nil
}

// DeletePath removes a file from the index
func DeletePath(relPath string) error {
	if DB == nil {
		return nil
	}

	if parts := parseInvoicePath(relPath); parts != nil {
		_, err := DB.Exec("DELETE FROM invoices WHERE id = ?", parts.ID)
		return err
	}

	if parts := parseHourPath(relPath); parts != nil {
		_, err := DB.Exec("DELETE FROM hours WHERE id = ?", parts.ID)
		return err
	}

	return nil
}

// MovePath handles file moves (delete old, add new)
func MovePath(dbPath, fromPath, toPath string) error {
	if err := DeletePath(fromPath); err != nil {
		return err
	}
	return SyncPath(dbPath, toPath)
}

func syncInvoice(p *PathParts) error {
	file, err := os.Open(p.FullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File deleted, remove from index
			_, err := DB.Exec("DELETE FROM invoices WHERE id = ?", p.ID)
			return err
		}
		return err
	}
	defer file.Close()

	var inv model.Invoice
	buf := bufio.NewReader(file)
	if _, err := toml.DecodeReader(buf, &inv); err != nil {
		return fmt.Errorf("idx: decode invoice %s: %w", p.FullPath, err)
	}

	// Determine tax category from Notes
	taxCat := deriveTaxCategory(&inv)

	_, err = DB.Exec(`
		INSERT OR REPLACE INTO invoices
		(id, entity, year, quarter, status, customer_name, customer_vat, invoiceid,
		 issuedate, duedate, paydate, tax_category, total_ex, total_tax, total_inc,
		 hour_file, notes, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Entity, p.Year, p.Quarter, p.Status,
		inv.Customer.Name, inv.Customer.Vat, inv.Meta.Invoiceid,
		inv.Meta.Issuedate, inv.Meta.Duedate, inv.Meta.Paydate, taxCat,
		inv.Total.Ex, inv.Total.Tax, inv.Total.Total,
		inv.Meta.HourFile, inv.Notes, time.Now().Format(time.RFC3339),
	)

	if config.Verbose && err == nil {
		fmt.Printf("idx: synced invoice %s (entity=%s, year=%d, Q%d, status=%s, tax=%s)\n",
			p.ID, p.Entity, p.Year, p.Quarter, p.Status, taxCat)
	}

	return err
}

// extractIssuedateFromFilename extracts YYYY-MM-DD from filename like h-2025-12.toml
func extractIssuedateFromFilename(filename string) string {
	// Pattern: h-YYYY-MM or similar
	re := regexp.MustCompile(`(\d{4})-(\d{2})`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) == 3 {
		return matches[1] + "-" + matches[2] + "-01"
	}
	return ""
}

func syncHour(p *PathParts) error {
	file, err := os.Open(p.FullPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := DB.Exec("DELETE FROM hours WHERE id = ?", p.ID)
			return err
		}
		return err
	}
	defer file.Close()

	var h model.Hour
	buf := bufio.NewReader(file)
	if _, err := toml.DecodeReader(buf, &h); err != nil {
		return fmt.Errorf("idx: decode hour %s: %w", p.FullPath, err)
	}

	// Use Issuedate from TOML if set, otherwise extract from filename
	issuedate := h.Issuedate
	if issuedate == "" {
		issuedate = extractIssuedateFromFilename(p.ID)
	}

	_, err = DB.Exec(`
		INSERT OR REPLACE INTO hours
		(id, entity, year, quarter, status, project, name, total_hours, hour_rate, issuedate, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Entity, p.Year, p.Quarter, p.Status,
		h.Project, h.Name, h.Total, h.HourRate, issuedate,
		time.Now().Format(time.RFC3339),
	)

	if config.Verbose && err == nil {
		fmt.Printf("idx: synced hour %s (entity=%s, year=%d, Q%d, issuedate=%s)\n",
			p.ID, p.Entity, p.Year, p.Quarter, issuedate)
	}

	return err
}

// deriveTaxCategory determines tax category from invoice Notes field
// Returns: "NL", "EU0", "WORLD0"
func deriveTaxCategory(inv *model.Invoice) string {
	if strings.Contains(inv.Notes, "Export") {
		return "WORLD0"
	}
	if strings.Contains(inv.Notes, "VAT Reverse charge") {
		return "EU0"
	}
	return "NL"
}
