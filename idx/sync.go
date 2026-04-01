package idx

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/model"
	"github.com/mpdroog/invoiced/purchase"
)

// debtorTOML is a local struct for decoding debtors.toml (avoids import cycle with entities)
type debtorTOML struct {
	Name         string
	Street1      string
	Street2      string
	VAT          string
	COC          string
	TAX          string
	NoteAdd      string
	BillingEmail []string
}

// projectTOML is a local struct for decoding projects.toml (avoids import cycle with entities)
type projectTOML struct {
	Name         string
	Debtor       string
	BillingEmail []string
	NoteAdd      string
	HourRate     float64
	DueDays      int
	PO           string
	Street1      string
}

// Status constants
const (
	statusConcept = "CONCEPT"
	statusPaid    = "PAID"
	statusUnpaid  = "UNPAID"
	bucketConcept = "concepts"
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

// Patterns for matching invoice, hour, purchase, debtor, and project paths
// e.g., "rootdev/2024/Q1/sales-invoices-paid/abc123.toml"
var (
	invoicePathRe  = regexp.MustCompile(`^([^/]+)/(\d{4})/(Q[1-4]|concepts)/(sales-invoices[^/]*)/([^/]+)\.toml$`)
	hourPathRe     = regexp.MustCompile(`^([^/]+)/(\d{4})/(Q[1-4]|concepts)/hours/([^/]+)\.toml$`)
	purchasePathRe = regexp.MustCompile(`^([^/]+)/(\d{4})/(Q[1-4])/(purchase-invoices[^/]*)/([^/]+)\.toml$`)
	debtorPathRe   = regexp.MustCompile(`^([^/]+)/debtors\.toml$`)
	projectPathRe  = regexp.MustCompile(`^([^/]+)/projects\.toml$`)
)

// parseInvoicePath extracts components from an invoice path
func parseInvoicePath(relPath string) *PathParts {
	matches := invoicePathRe.FindStringSubmatch(relPath)
	if matches == nil {
		return nil
	}

	year, _ := strconv.Atoi(matches[2])
	quarter := 0
	if matches[3] != bucketConcept {
		q := matches[3][1] - '0' // Q1 -> 1
		quarter = int(q)
	}

	// Determine status from bucket name
	bucket := matches[4]
	status := statusConcept
	if strings.Contains(bucket, "-paid") {
		status = statusPaid
	} else if strings.Contains(bucket, "-unpaid") {
		status = statusUnpaid
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
	status := statusConcept
	if matches[3] != bucketConcept {
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

// parsePurchasePath extracts components from a purchase invoice path
func parsePurchasePath(relPath string) *PathParts {
	matches := purchasePathRe.FindStringSubmatch(relPath)
	if matches == nil {
		return nil
	}

	year, _ := strconv.Atoi(matches[2])
	q := matches[3][1] - '0' // Q1 -> 1
	quarter := int(q)

	// Determine status from bucket name
	bucket := matches[4]
	status := statusUnpaid
	if strings.Contains(bucket, "-paid") {
		status = statusPaid
	}

	return &PathParts{
		ID:      matches[5],
		Entity:  matches[1],
		Year:    year,
		Quarter: quarter,
		Status:  status,
		Type:    "purchase",
	}
}

// parseDebtorPath extracts entity from a debtors.toml path
func parseDebtorPath(relPath string) string {
	matches := debtorPathRe.FindStringSubmatch(relPath)
	if matches == nil {
		return ""
	}
	return matches[1]
}

// parseProjectPath extracts entity from a projects.toml path
func parseProjectPath(relPath string) string {
	matches := projectPathRe.FindStringSubmatch(relPath)
	if matches == nil {
		return ""
	}
	return matches[1]
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

	// Try purchase invoice
	if parts := parsePurchasePath(relPath); parts != nil {
		parts.FullPath = fullPath
		return syncPurchase(parts)
	}

	// Try debtors
	if entity := parseDebtorPath(relPath); entity != "" {
		return syncDebtors(fullPath, entity)
	}

	// Try projects
	if entity := parseProjectPath(relPath); entity != "" {
		return syncProjects(fullPath, entity)
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
		_, err := DB.ExecContext(context.Background(), "DELETE FROM invoices WHERE id = ?", parts.ID)
		return err
	}

	if parts := parseHourPath(relPath); parts != nil {
		_, err := DB.ExecContext(context.Background(), "DELETE FROM hours WHERE id = ?", parts.ID)
		return err
	}

	if parts := parsePurchasePath(relPath); parts != nil {
		_, err := DB.ExecContext(context.Background(), "DELETE FROM purchase_invoices WHERE id = ?", parts.ID)
		return err
	}

	if entity := parseDebtorPath(relPath); entity != "" {
		_, err := DB.ExecContext(context.Background(), "DELETE FROM debtors WHERE entity = ?", entity)
		return err
	}

	if entity := parseProjectPath(relPath); entity != "" {
		_, err := DB.ExecContext(context.Background(), "DELETE FROM projects WHERE entity = ?", entity)
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
			_, err := DB.ExecContext(context.Background(), "DELETE FROM invoices WHERE id = ?", p.ID)
			return err
		}
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var inv model.Invoice
	buf := bufio.NewReader(file)
	if _, err := toml.NewDecoder(buf).Decode(&inv); err != nil {
		return fmt.Errorf("idx: decode invoice %s: %w", p.FullPath, err)
	}

	// Determine tax category from Notes
	taxCat := deriveTaxCategory(&inv)

	_, err = DB.ExecContext(context.Background(), `
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
			_, err := DB.ExecContext(context.Background(), "DELETE FROM hours WHERE id = ?", p.ID)
			return err
		}
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var h model.Hour
	buf := bufio.NewReader(file)
	if _, err := toml.NewDecoder(buf).Decode(&h); err != nil {
		return fmt.Errorf("idx: decode hour %s: %w", p.FullPath, err)
	}

	// Use Issuedate from TOML if set, otherwise extract from filename
	issuedate := h.Issuedate
	if issuedate == "" {
		issuedate = extractIssuedateFromFilename(p.ID)
	}

	_, err = DB.ExecContext(context.Background(), `
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

func syncPurchase(p *PathParts) error {
	file, err := os.Open(p.FullPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := DB.ExecContext(context.Background(), "DELETE FROM purchase_invoices WHERE id = ?", p.ID)
			return err
		}
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var inv purchase.PurchaseInvoice
	buf := bufio.NewReader(file)
	if _, err := toml.NewDecoder(buf).Decode(&inv); err != nil {
		return fmt.Errorf("idx: decode purchase %s: %w", p.FullPath, err)
	}

	_, err = DB.ExecContext(context.Background(), `
		INSERT OR REPLACE INTO purchase_invoices
		(id, entity, year, quarter, status, supplier_name, supplier_vat, invoiceid,
		 issuedate, duedate, paydate, total_ex, total_tax, total_inc, currency,
		 payment_ref, iban, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Entity, p.Year, p.Quarter, p.Status,
		inv.Supplier.Name, inv.Supplier.VAT, inv.ID,
		inv.Issuedate, inv.Duedate, inv.Paydate,
		inv.TotalEx, inv.TotalTax, inv.TotalInc, inv.Currency,
		inv.PaymentRef, inv.IBAN, time.Now().Format(time.RFC3339),
	)

	if config.Verbose && err == nil {
		fmt.Printf("idx: synced purchase %s (entity=%s, year=%d, Q%d, status=%s)\n",
			p.ID, p.Entity, p.Year, p.Quarter, p.Status)
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

// syncDebtors syncs all debtors from a debtors.toml file
func syncDebtors(fullPath, entity string) error {
	file, err := os.Open(fullPath) //nolint:gosec // path from internal db, not user input
	if err != nil {
		if os.IsNotExist(err) {
			// File deleted, remove all debtors for this entity
			_, err := DB.ExecContext(context.Background(), "DELETE FROM debtors WHERE entity = ?", entity)
			return err
		}
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var debtorList map[string]debtorTOML
	buf := bufio.NewReader(file)
	if _, err := toml.NewDecoder(buf).Decode(&debtorList); err != nil {
		return fmt.Errorf("idx: decode debtors %s: %w", fullPath, err)
	}

	// Delete existing debtors for this entity and insert fresh
	if _, err := DB.ExecContext(context.Background(), "DELETE FROM debtors WHERE entity = ?", entity); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	for key, d := range debtorList {
		_, err = DB.ExecContext(context.Background(), `
			INSERT INTO debtors (id, entity, name, street1, street2, vat, coc, tax, note_add, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			entity+"/"+key, entity, d.Name, d.Street1, d.Street2, d.VAT, d.COC, d.TAX, d.NoteAdd, now,
		)
		if err != nil {
			return err
		}
	}

	if config.Verbose {
		fmt.Printf("idx: synced %d debtors for entity %s\n", len(debtorList), entity)
	}

	return nil
}

// syncProjects syncs all projects from a projects.toml file
func syncProjects(fullPath, entity string) error {
	file, err := os.Open(fullPath) //nolint:gosec // path from internal db, not user input
	if err != nil {
		if os.IsNotExist(err) {
			// File deleted, remove all projects for this entity
			_, err := DB.ExecContext(context.Background(), "DELETE FROM projects WHERE entity = ?", entity)
			return err
		}
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	var projectList map[string]projectTOML
	buf := bufio.NewReader(file)
	if _, err := toml.NewDecoder(buf).Decode(&projectList); err != nil {
		return fmt.Errorf("idx: decode projects %s: %w", fullPath, err)
	}

	// Delete existing projects for this entity and insert fresh
	if _, err := DB.ExecContext(context.Background(), "DELETE FROM projects WHERE entity = ?", entity); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	for key, p := range projectList {
		_, err = DB.ExecContext(context.Background(), `
			INSERT INTO projects (id, entity, name, debtor, hour_rate, due_days, po, street1, note_add, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			entity+"/"+key, entity, p.Name, p.Debtor, p.HourRate, p.DueDays, p.PO, p.Street1, p.NoteAdd, now,
		)
		if err != nil {
			return err
		}
	}

	if config.Verbose {
		fmt.Printf("idx: synced %d projects for entity %s\n", len(projectList), entity)
	}

	return nil
}
