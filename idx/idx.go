// Package idx provides a SQLite read cache for fast aggregation queries.
// TOML files remain the source of truth; SQLite is rebuilt on startup
// and kept in sync via hooks in db.Update().
package idx

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/mpdroog/invoiced/config"
	_ "modernc.org/sqlite" // SQLite driver
)

// DB is the SQLite database connection for the search index.
var DB *sql.DB

// Open initializes the SQLite index at dbPath/index.db
func Open(dbPath string) error {
	indexPath := filepath.Join(dbPath, "index.db")

	db, err := sql.Open("sqlite", indexPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return err
	}
	DB = db

	// Enable foreign keys and optimize for read-heavy workload
	if _, err := DB.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`); err != nil {
		return err
	}

	return createTables()
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// IsEmpty returns true if the invoices table has no rows
func IsEmpty() bool {
	var count int
	err := DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM invoices").Scan(&count)
	if err != nil {
		// Table might not exist yet
		return true
	}
	return count == 0
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metadata (
		key   TEXT PRIMARY KEY,
		value TEXT
	);

	CREATE TABLE IF NOT EXISTS invoices (
		id            TEXT PRIMARY KEY,  -- conceptid (filename without .toml)
		entity        TEXT NOT NULL,
		year          INTEGER NOT NULL,
		quarter       INTEGER NOT NULL,  -- 1-4
		status        TEXT NOT NULL,     -- CONCEPT, UNPAID, PAID

		-- Indexed fields for queries
		customer_name TEXT,
		customer_vat  TEXT,
		invoiceid     TEXT,              -- Official invoice number
		issuedate     TEXT,
		duedate       TEXT,
		paydate       TEXT,

		-- Tax categorization (derived from Notes + Customer.TAX)
		tax_category  TEXT,              -- NL, EU0, WORLD0

		-- Totals for aggregations (stored as TEXT for decimal precision)
		total_ex      TEXT,
		total_tax     TEXT,
		total_inc     TEXT,

		-- Metadata
		hour_file     TEXT,
		notes         TEXT,
		updated_at    TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_invoices_entity_year ON invoices(entity, year);
	CREATE INDEX IF NOT EXISTS idx_invoices_quarter ON invoices(entity, year, quarter, status);
	CREATE INDEX IF NOT EXISTS idx_invoices_tax ON invoices(entity, year, quarter, tax_category);
	CREATE INDEX IF NOT EXISTS idx_invoices_customer ON invoices(entity, customer_name);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_invoiceid ON invoices(entity, invoiceid);

	CREATE TABLE IF NOT EXISTS hours (
		id            TEXT PRIMARY KEY,
		entity        TEXT NOT NULL,
		year          INTEGER NOT NULL,
		quarter       INTEGER NOT NULL,
		status        TEXT NOT NULL,

		project       TEXT,
		name          TEXT,
		total_hours   TEXT,
		hour_rate     REAL,
		issuedate     TEXT,  -- YYYY-MM-DD

		updated_at    TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_hours_entity_year ON hours(entity, year);
	CREATE INDEX IF NOT EXISTS idx_hours_quarter ON hours(entity, year, quarter);
	CREATE INDEX IF NOT EXISTS idx_hours_issuedate ON hours(entity, issuedate);

	CREATE TABLE IF NOT EXISTS purchase_invoices (
		id            TEXT PRIMARY KEY,  -- sanitized filename
		entity        TEXT NOT NULL,
		year          INTEGER NOT NULL,
		quarter       INTEGER NOT NULL,
		status        TEXT NOT NULL,     -- UNPAID, PAID

		-- Supplier info
		supplier_name TEXT,
		supplier_vat  TEXT,

		-- Invoice details
		invoiceid     TEXT,              -- Original invoice ID from supplier
		issuedate     TEXT,
		duedate       TEXT,
		paydate       TEXT,

		-- Totals
		total_ex      TEXT,
		total_tax     TEXT,
		total_inc     TEXT,
		currency      TEXT,

		-- Payment info
		payment_ref   TEXT,
		iban          TEXT,

		updated_at    TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_purchases_entity_year ON purchase_invoices(entity, year);
	CREATE INDEX IF NOT EXISTS idx_purchases_quarter ON purchase_invoices(entity, year, quarter, status);
	CREATE INDEX IF NOT EXISTS idx_purchases_supplier ON purchase_invoices(entity, supplier_name);

	CREATE TABLE IF NOT EXISTS debtors (
		id              TEXT PRIMARY KEY,  -- key (slug) from TOML
		entity          TEXT NOT NULL,
		name            TEXT,
		street1         TEXT,
		street2         TEXT,
		vat             TEXT,
		coc             TEXT,
		tax             TEXT,              -- NL21, EU0, WORLD0
		note_add        TEXT,
		accounting_code TEXT,              -- Relation code for accounting software
		updated_at      TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_debtors_entity ON debtors(entity);
	CREATE INDEX IF NOT EXISTS idx_debtors_name ON debtors(entity, name);

	CREATE TABLE IF NOT EXISTS projects (
		id            TEXT PRIMARY KEY,  -- key (slug) from TOML
		entity        TEXT NOT NULL,
		name          TEXT,
		debtor        TEXT,              -- references debtor key
		hour_rate     REAL,
		due_days      INTEGER,
		po            TEXT,
		street1       TEXT,
		note_add      TEXT,
		updated_at    TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_projects_entity ON projects(entity);
	CREATE INDEX IF NOT EXISTS idx_projects_debtor ON projects(entity, debtor);
	`

	_, err := DB.ExecContext(context.Background(), schema)
	if err != nil {
		return err
	}

	if config.Verbose {
		log.Printf("idx: tables created/verified")
	}
	return nil
}

// DeleteIndex removes the index.db file (for full rebuild)
func DeleteIndex(dbPath string) error {
	indexPath := filepath.Join(dbPath, "index.db")
	if err := os.Remove(indexPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	// Also remove WAL files
	if err := os.Remove(indexPath + "-wal"); err != nil && !os.IsNotExist(err) {
		log.Printf("idx.DeleteIndex wal: %s", err)
	}
	if err := os.Remove(indexPath + "-shm"); err != nil && !os.IsNotExist(err) {
		log.Printf("idx.DeleteIndex shm: %s", err)
	}
	return nil
}
