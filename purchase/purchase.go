// Package purchase handles purchase invoice management.
package purchase

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/utils"
	"github.com/mpdroog/invoiced/writer"
)

// List returns all purchase invoices for a year.
func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	args := r.URL.Query()

	paths := db.PurchaseListPaths(entity, year)

	from, e := strconv.Atoi(args.Get("from"))
	if e != nil {
		httputil.InternalError(w, "purchase.List from", e)
		return
	}
	count, e := strconv.Atoi(args.Get("count"))
	if e != nil {
		httputil.InternalError(w, "purchase.List count", e)
		return
	}

	list := make(map[string][]*PurchaseInvoice)
	mem := new(PurchaseInvoice)

	e = db.View(func(t *db.Txn) error {
		_, e := t.List(paths, db.Pagination{From: from, Count: count}, &mem, func(_, _, fpath string) error {
			list[fpath] = append(list[fpath], mem)
			mem = new(PurchaseInvoice)
			return nil
		})
		return e
	})
	if e != nil {
		httputil.BadRequest(w, "purchase.List", e)
		return
	}

	res := &ListReply{
		Invoices: list,
	}

	if config.Verbose {
		log.Printf("purchase.List count=%d", len(list))
	}
	if e := writer.Encode(w, r, res); e != nil {
		httputil.LogErr("purchase.List", e)
	}
}

// Load retrieves a single purchase invoice by ID.
func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	bucket := ps.ByName("bucket")
	if config.Verbose {
		log.Printf("purchase.Load with id=%s", name)
	}

	paths := []string{
		db.PurchasePath(entity, year, bucket, name, true),
		db.PurchasePath(entity, year, bucket, name, false),
	}

	u := new(PurchaseInvoice)
	e := db.View(func(t *db.Txn) error {
		return t.OpenFirst(paths, u)
	})
	if e != nil {
		httputil.BadRequest(w, "purchase.Load", e)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("purchase.Load", e)
	}
}

// Upload handles UBL XML upload and extracts the embedded PDF
func Upload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if err := r.ParseMultipartForm(32 << 20); err != nil { //nolint:gosec // G120: 32MB limit is set
		http.Error(w, "Failed to parse form", 400)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file", 400)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xml") {
		http.Error(w, "Only XML files accepted", 400)
		return
	}

	// Parse UBL XML
	inv, pdfData, err := ParseUBL(file)
	if err != nil {
		log.Printf("purchase.Upload parse error: %q", err.Error())
		http.Error(w, fmt.Sprintf("Failed to parse XML: %q", err.Error()), 400)
		return
	}

	inv.XMLFilename = header.Filename

	// Determine quarter from issue date
	issueDate, err := time.Parse("2006-01-02", inv.Issuedate)
	if err != nil {
		http.Error(w, "Invalid issue date in XML", 400)
		return
	}
	quarter := fmt.Sprintf("Q%d", utils.YearQuarter(issueDate))

	// Use supplier name + invoice ID as unique identifier
	safeID := sanitizeFilename(inv.Supplier.Name + "-" + inv.ID)

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Upload purchase invoice %s from %s", inv.ID, inv.Supplier.Name),
	}

	e := db.Update(change, func(t *db.Txn) error {
		basePath := fmt.Sprintf("%s/%s/%s/purchase-invoices-unpaid", entity, year, quarter)

		// Save TOML
		tomlPath := fmt.Sprintf("%s/%s.toml", basePath, safeID)
		if err := t.Save(tomlPath, true, inv); err != nil {
			return err
		}

		// Save PDF if embedded
		if len(pdfData) > 0 {
			pdfPath := fmt.Sprintf("%s/%s.pdf", basePath, safeID)
			if err := t.SaveRaw(pdfPath, bytes.NewReader(pdfData)); err != nil {
				return err
			}
		}

		return nil
	})

	if e != nil {
		httputil.InternalError(w, "purchase.Upload", e)
		return
	}

	if e := writer.Encode(w, r, inv); e != nil {
		httputil.LogErr("purchase.Upload", e)
	}
}

// Paid marks a purchase invoice as paid
func Paid(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if config.Verbose {
		log.Printf("purchase.Paid with id=%s", name)
	}

	from := db.PurchasePath(entity, year, bucket, name, false)
	to := db.PurchasePath(entity, year, bucket, name, true)
	u := new(PurchaseInvoice)

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Mark purchase invoice %s as paid", name),
	}

	e := db.Update(change, func(t *db.Txn) error {
		if e := t.Open(from, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		u.Status = "PAID"
		u.Paydate = time.Now().Format("2006-01-02")
		if e := t.Save(to, true, u); e != nil {
			return fmt.Errorf("save: %w", e)
		}
		if e := t.Remove(from); e != nil {
			return fmt.Errorf("remove: %w", e)
		}

		// Also move PDF if exists
		pdfFrom := strings.TrimSuffix(from, ".toml") + ".pdf"
		pdfTo := strings.TrimSuffix(to, ".toml") + ".pdf"
		_ = t.Move(pdfFrom, pdfTo) // Ignore error if PDF doesn't exist

		return nil
	})

	if e != nil {
		httputil.BadRequest(w, "purchase.Paid", e)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("purchase.Paid", e)
	}
}

// PDF serves the embedded PDF file
func PDF(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	// PDF paths (same as TOML but with .pdf extension)
	paths := []string{
		strings.TrimSuffix(db.PurchasePath(entity, year, bucket, name, true), ".toml") + ".pdf",
		strings.TrimSuffix(db.PurchasePath(entity, year, bucket, name, false), ".toml") + ".pdf",
	}

	e := db.View(func(t *db.Txn) error {
		for _, path := range paths {
			f, err := t.OpenRaw(path)
			if err != nil {
				continue
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Printf("close: %s", err)
				}
			}()

			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, name))

			buf := make([]byte, 32*1024)
			for {
				n, err := f.Read(buf)
				if n > 0 {
					if _, werr := w.Write(buf[:n]); werr != nil {
						log.Printf("purchase.PDF write: %s", werr)
						break
					}
				}
				if err != nil {
					break
				}
			}
			return nil
		}
		return fmt.Errorf("PDF not found")
	})

	if e != nil {
		httputil.NotFound(w, "purchase.PDF", e)
	}
}

// Delete removes a purchase invoice
func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if name == "" {
		http.Error(w, "Please supply an id to delete", 400)
		return
	}

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Delete purchase invoice %s", name),
	}

	e := db.Update(change, func(t *db.Txn) error {
		// Try both paid and unpaid locations
		paths := []string{
			db.PurchasePath(entity, year, bucket, name, true),
			db.PurchasePath(entity, year, bucket, name, false),
		}

		for _, path := range paths {
			if err := t.Remove(path); err == nil {
				// Also remove PDF
				pdfPath := strings.TrimSuffix(path, ".toml") + ".pdf"
				_ = t.Remove(pdfPath)
				return nil
			}
		}
		return fmt.Errorf("purchase invoice not found")
	})

	if e != nil {
		httputil.InternalError(w, "purchase.Delete", e)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"ok": true}`)); err != nil {
		log.Printf("purchase.Delete write: %s", err)
	}
}

func sanitizeFilename(s string) string {
	// Remove/replace characters that are problematic in filenames
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
		" ", "-",
	)
	result := replacer.Replace(s)
	// Remove consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.ToLower(result)
}
