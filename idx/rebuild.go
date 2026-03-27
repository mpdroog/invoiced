package idx

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mpdroog/invoiced/config"
)

// Rebuild clears the index and rebuilds from all TOML files
func Rebuild(dbPath string) error {
	if DB == nil {
		return nil
	}

	log.Printf("idx: rebuilding index from %s", dbPath)

	// Clear existing data
	if _, err := DB.Exec("DELETE FROM invoices"); err != nil {
		return err
	}
	if _, err := DB.Exec("DELETE FROM hours"); err != nil {
		return err
	}

	var invoiceCount, hourCount int

	// Walk all TOML files
	err := filepath.Walk(dbPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-TOML files
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".toml") {
			return nil
		}

		// Skip system files
		name := info.Name()
		if name == ".DS_Store" || name == "entities.toml" ||
			name == "debtors.toml" || name == "projects.toml" ||
			name == "counters.toml" {
			return nil
		}

		// Get path relative to dbPath
		relPath, err := filepath.Rel(dbPath, path)
		if err != nil {
			return err
		}

		// Determine type and sync
		if parts := parseInvoicePath(relPath); parts != nil {
			parts.FullPath = path
			if err := syncInvoice(parts); err != nil {
				log.Printf("idx: error syncing invoice %s: %v", relPath, err)
				// Continue with other files
			} else {
				invoiceCount++
			}
		} else if parts := parseHourPath(relPath); parts != nil {
			parts.FullPath = path
			if err := syncHour(parts); err != nil {
				log.Printf("idx: error syncing hour %s: %v", relPath, err)
			} else {
				hourCount++
			}
		} else if config.Verbose {
			// Only log unprocessed files in verbose mode
			if strings.Contains(relPath, "/sales-invoices") || strings.Contains(relPath, "/hours/") {
				log.Printf("idx: unprocessed file %s", relPath)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("idx: rebuilt index with %d invoices and %d hours", invoiceCount, hourCount)
	return nil
}
