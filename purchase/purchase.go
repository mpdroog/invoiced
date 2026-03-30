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
	"github.com/mpdroog/invoiced/utils"
	"github.com/mpdroog/invoiced/writer"
)

func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	args := r.URL.Query()

	paths := []string{
		fmt.Sprintf("%s/%s/{all}/purchase-invoices-unpaid", entity, year),
		fmt.Sprintf("%s/%s/{all}/purchase-invoices-paid", entity, year),
	}

	from, e := strconv.Atoi(args.Get("from"))
	if e != nil {
		log.Printf("purchase.List from: %s", e.Error())
		http.Error(w, "purchase.List fail", http.StatusInternalServerError)
		return
	}
	count, e := strconv.Atoi(args.Get("count"))
	if e != nil {
		log.Printf("purchase.List count: %s", e.Error())
		http.Error(w, "purchase.List fail", http.StatusInternalServerError)
		return
	}

	list := make(map[string][]*PurchaseInvoice)
	mem := new(PurchaseInvoice)

	e = db.View(func(t *db.Txn) error {
		_, e := t.List(paths, db.Pagination{From: from, Count: count}, &mem, func(filename, filepath, fpath string) error {
			list[fpath] = append(list[fpath], mem)
			mem = new(PurchaseInvoice)
			return nil
		})
		return e
	})
	if e != nil {
		log.Printf("purchase.List %s", e.Error())
		http.Error(w, fmt.Sprintf("purchase.List failed scanning disk"), 400)
		return
	}

	res := &ListReply{
		Invoices: list,
	}

	if config.Verbose {
		log.Printf("purchase.List count=%d", len(list))
	}
	if e := writer.Encode(w, r, res); e != nil {
		log.Printf("purchase.List %s", e.Error())
	}
}

func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	bucket := ps.ByName("bucket")
	log.Printf("purchase.Load with id=%s", name)

	paths := []string{
		fmt.Sprintf("%s/%s/%s/purchase-invoices-paid/%s.toml", entity, year, bucket, name),
		fmt.Sprintf("%s/%s/%s/purchase-invoices-unpaid/%s.toml", entity, year, bucket, name),
	}

	u := new(PurchaseInvoice)
	e := db.View(func(t *db.Txn) error {
		return t.OpenFirst(paths, u)
	})
	if e != nil {
		log.Printf("purchase.Load %s", e.Error())
		http.Error(w, fmt.Sprintf("purchase.Load failed loading file from disk"), 400)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("purchase.Load %s", e.Error())
	}
}

// Upload handles UBL XML upload and extracts the embedded PDF
func Upload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "Failed to parse form", 400)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file", 400)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xml") {
		http.Error(w, "Only XML files accepted", 400)
		return
	}

	// Parse UBL XML
	inv, pdfData, err := ParseUBL(file)
	if err != nil {
		log.Printf("purchase.Upload parse error: %s", err.Error())
		http.Error(w, fmt.Sprintf("Failed to parse XML: %s", err.Error()), 400)
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
		log.Printf("purchase.Upload %s", e.Error())
		http.Error(w, "Failed to save purchase invoice", 500)
		return
	}

	if e := writer.Encode(w, r, inv); e != nil {
		log.Printf("purchase.Upload %s", e.Error())
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

	from := fmt.Sprintf("%s/%s/%s/purchase-invoices-unpaid/%s.toml", entity, year, bucket, name)
	to := fmt.Sprintf("%s/%s/%s/purchase-invoices-paid/%s.toml", entity, year, bucket, name)
	u := new(PurchaseInvoice)

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Mark purchase invoice %s as paid", name),
	}

	e := db.Update(change, func(t *db.Txn) error {
		if e := t.Open(from, u); e != nil {
			return fmt.Errorf("Paid::Open %v", e)
		}
		u.Status = "PAID"
		u.Paydate = time.Now().Format("2006-01-02")
		if e := t.Save(to, true, u); e != nil {
			return fmt.Errorf("Paid::Save %v", e)
		}
		if e := t.Remove(from); e != nil {
			return fmt.Errorf("Paid::Remove %v", e)
		}

		// Also move PDF if exists
		pdfFrom := strings.TrimSuffix(from, ".toml") + ".pdf"
		pdfTo := strings.TrimSuffix(to, ".toml") + ".pdf"
		_ = t.Move(pdfFrom, pdfTo) // Ignore error if PDF doesn't exist

		return nil
	})

	if e != nil {
		log.Printf("purchase.Paid %s", e.Error())
		http.Error(w, "Failed to mark as paid", 400)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("purchase.Paid %s", e.Error())
	}
}

// PDF serves the embedded PDF file
func PDF(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	paths := []string{
		fmt.Sprintf("%s/%s/%s/purchase-invoices-paid/%s.pdf", entity, year, bucket, name),
		fmt.Sprintf("%s/%s/%s/purchase-invoices-unpaid/%s.pdf", entity, year, bucket, name),
	}

	e := db.View(func(t *db.Txn) error {
		for _, path := range paths {
			f, err := t.OpenRaw(path)
			if err != nil {
				continue
			}
			defer f.Close()

			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, name))

			buf := make([]byte, 32*1024)
			for {
				n, err := f.Read(buf)
				if n > 0 {
					w.Write(buf[:n])
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
		log.Printf("purchase.PDF %s", e.Error())
		http.Error(w, "PDF not found", 404)
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
			fmt.Sprintf("%s/%s/%s/purchase-invoices-paid/%s.toml", entity, year, bucket, name),
			fmt.Sprintf("%s/%s/%s/purchase-invoices-unpaid/%s.toml", entity, year, bucket, name),
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
		log.Printf("purchase.Delete %s", e.Error())
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok": true}`))
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
