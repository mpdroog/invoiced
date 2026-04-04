package invoice

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/jung-kurt/gofpdf"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/utils"
	"github.com/mpdroog/invoiced/writer"
	"gopkg.in/validator.v2"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// EUTaxComment is the VAT reverse charge text for EU invoices.
const EUTaxComment = "VAT Reverse charge"

// WorldTaxComment is the export text for non-EU invoices.
const WorldTaxComment = "Export"

// InputError represents validation errors for user input.
type InputError struct {
	Error  string
	Fields validator.ErrorMap
}

// Note: rand.Seed is no longer needed as of Go 1.20

// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func randStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))] //nolint:gosec // G404: used for temp filenames, not crypto
	}
	return string(b)
}

// Delete removes a concept invoice.
func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to delete", http.StatusBadRequest)
		return
	}
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, db.ActionDelete, db.ResourceInvoice, name, "concept"),
	}
	e := db.Update(&change, func(t *db.Txn) error {
		path := db.ConceptInvoicePath(entity, year, name)
		u := new(Invoice)
		if e := t.Open(path, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		if len(u.Meta.Invoiceid) > 0 {
			return fmt.Errorf("reject deleting finalized invoice (this will break your accounting history)")
		}
		return t.Remove(path)
	})
	if e != nil {
		httputil.InternalError(w, "invoice.Delete", e)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, e := w.Write([]byte(`{"ok": true}`)); e != nil {
		httputil.LogErr("invoice.Delete", e)
	}
}

func vatCountry(nr string) string {
	country := "nl"
	if len(nr) > 3 {
		country = strings.ToLower(nr[0:2])
	}
	return country
}

// Finalize locks an invoice for changes and assigns an invoice ID.
func Finalize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to finalize", http.StatusBadRequest)
		return
	}
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Finalize with conceptid=%s", name)
	}

	var u *Invoice
	bucketTo := ""
	change := db.Commit{
		Name:  r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		// Message set inside transaction after invoice ID is assigned
	}
	e := db.Update(&change, func(t *db.Txn) error {
		from := db.ConceptInvoicePath(entity, year, name)
		u = new(Invoice)
		if e := t.Open(from, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		if len(u.Meta.Issuedate) == 0 {
			u.Meta.Issuedate = time.Now().Format("2006-01-02")
		}
		u.Meta.Status = "FINAL"

		if u.Meta.Invoiceid == "" {
			idx, e := NextInvoiceID(entity, t)
			if e != nil {
				return fmt.Errorf("next invoice id: %w", e)
			}
			u.Meta.Invoiceid = utils.CreateInvoiceID(time.Now(), idx)
			u.Mail.Subject += " #" + u.Meta.Invoiceid
			if config.Verbose {
				log.Printf("invoice.Finalize create conceptId=%s invoiceId=%s", u.Meta.Conceptid, u.Meta.Invoiceid)
			}

			// Set commit message now that we have the invoice ID
			change.Message = db.FormatCommitMsg(entity, db.ActionFinalize, db.ResourceInvoice, name, "->", u.Meta.Invoiceid)

			if u.Customer.Tax != "" {
				// If outside NL we add a special comment above the invoice so it gets administrated correctly
				if u.Customer.Tax == "EU0" && !strings.Contains(u.Notes, EUTaxComment) {
					if len(u.Notes) > 0 {
						u.Notes += "\n\n"
					}
					u.Notes += EUTaxComment
				} else if u.Customer.Tax == "WORLD0" && !strings.Contains(u.Notes, WorldTaxComment) {
					if len(u.Notes) > 0 {
						u.Notes += "\n\n"
					}
					u.Notes += WorldTaxComment
				}
			} else {
				// old style for older invoices (before 2024Q3)
				if vatCountry(u.Customer.Vat) != "nl" && !strings.Contains(u.Notes, EUTaxComment) {
					if len(u.Notes) > 0 {
						u.Notes += "\n\n"
					}
					u.Notes += EUTaxComment
				}
			}
		} else {
			// Re-finalize: invoice already has an ID
			change.Message = db.FormatCommitMsg(entity, db.ActionFinalize, db.ResourceInvoice, name, "->", u.Meta.Invoiceid)
		}

		now, e := time.Parse("2006-01-02", u.Meta.Issuedate)
		if e != nil {
			return fmt.Errorf("parse issue date: %w", e)
		}
		bucketTo = fmt.Sprintf("Q%d", utils.YearQuarter(now))
		to := db.InvoicePath(entity, year, bucketTo, name, false)
		if e := t.Save(to, true, u); e != nil {
			return fmt.Errorf("save: %w", e)
		}
		if e := t.Remove(from); e != nil {
			return fmt.Errorf("remove: %w", e)
		}
		return nil
	})
	if e != nil {
		httputil.InternalError(w, "invoice.Finalize", e)
		return
	}

	w.Header().Set("X-Bucket-Change", bucketTo)
	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("invoice.Finalize", e)
	}
}

// Reset moves a finalized invoice back to concept status.
func Reset(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Reset with conceptid=%s", name)
	}

	from := db.InvoicePath(entity, year, bucket, name, false)
	bucketTo := "concepts"
	u := new(Invoice)

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, db.ActionUnpay, db.ResourceInvoice, name, "reset to concept"),
	}
	e := db.Update(&change, func(t *db.Txn) error {
		if e := t.Open(from, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		to := db.ConceptInvoicePath(entity, year, u.Meta.Conceptid)
		u.Meta.Status = "CONCEPT"
		if e := t.Save(to, true, u); e != nil {
			return fmt.Errorf("save: %w", e)
		}
		if e := t.Remove(from); e != nil {
			return fmt.Errorf("remove: %w", e)
		}
		return nil
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.Reset", e)
		return
	}

	w.Header().Set("X-Bucket-Change", bucketTo)
	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("invoice.Reset", e)
	}
}

// Paid marks an invoice as paid and moves it to the paid folder.
func Paid(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Paid with conceptid=%s", name)
	}

	from := db.InvoicePath(entity, year, bucket, name, false)
	to := db.InvoicePath(entity, year, bucket, name, true)
	u := new(Invoice)

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, db.ActionPay, db.ResourceInvoice, name),
	}
	e := db.Update(&change, func(t *db.Txn) error {
		if e := t.Open(from, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		u.Meta.Paydate = time.Now().Format("2006-01-02")
		if e := t.Save(to, true, u); e != nil {
			return fmt.Errorf("save: %w", e)
		}
		if e := t.Remove(from); e != nil {
			return fmt.Errorf("remove: %w", e)
		}
		return nil
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.Paid", e)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("invoice.Paid", e)
	}
}

// Save creates or updates a concept invoice.
func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	u := new(Invoice)
	if e := writer.Decode(r, u); e != nil {
		httputil.BadRequest(w, "invoice.Save decode", e)
		return
	}
	if e := validator.Validate(u); e != nil {
		// Errors as JSON
		w.WriteHeader(http.StatusExpectationFailed)
		var errMap validator.ErrorMap
		if errors.As(e, &errMap) {
			input := InputError{
				Error:  "Input invalid",
				Fields: errMap,
			}
			if e := writer.Encode(w, r, input); e != nil {
				httputil.LogErr("invoice.Save encode", e)
			}
		} else {
			httputil.LogErr("invoice.Save validate", e)
		}
		return
	}

	isNew := u.Meta.Status == "NEW"

	if u.Meta.Conceptid == "" {
		u.Meta.Conceptid = randStringBytesRmndr(12)
		log.Printf("invoice.Save create conceptId=%s", u.Meta.Conceptid)
	} else {
		log.Printf("invoice.Save update conceptId=%s", u.Meta.Conceptid)
	}
	u.Meta.Status = "CONCEPT"

	action := db.ActionUpdate
	if isNew {
		action = db.ActionCreate
	}
	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, action, db.ResourceInvoice, u.Meta.Conceptid),
	}
	e := db.Update(&change, func(t *db.Txn) error {
		path := db.ConceptInvoicePath(entity, year, u.Meta.Conceptid)
		return t.Save(path, isNew, u)
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.Save", e)
		return
	}

	w.Header().Set("X-Bucket-Change", "concepts")
	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("invoice.Save", e)
	}
}

// Load retrieves a single invoice by ID.
func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	bucket := ps.ByName("bucket")
	if config.Verbose {
		log.Printf("invoice.Load with conceptid=%s", name)
	}

	var paths []string
	if bucket == "concepts" {
		paths = []string{db.ConceptInvoicePath(entity, year, name)}
	} else {
		paths = db.InvoiceSearchPaths(entity, year, bucket, name)
	}

	u := new(Invoice)
	e := db.View(func(t *db.Txn) error {
		return t.OpenFirst(paths, u)
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.Load", e)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("invoice.Load", e)
	}
}

// List returns all invoices for a year with pagination.
func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	args := r.URL.Query()

	paths := db.InvoiceListPaths(entity, year)

	from, e := strconv.Atoi(args.Get("from"))
	if e != nil {
		httputil.InternalError(w, "invoice.List from", e)
		return
	}
	count, e := strconv.Atoi(args.Get("count"))
	if e != nil {
		httputil.InternalError(w, "invoice.List count", e)
		return
	}

	list := make(map[string][]*Invoice)
	mem := new(Invoice)

	e = db.View(func(t *db.Txn) error {
		_, e := t.List(paths, db.Pagination{From: from, Count: count}, &mem, func(_, _, fpath string) error {
			list[fpath] = append(list[fpath], mem)
			mem = new(Invoice)
			return nil
		})
		return e
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.List", e)
		return
	}

	res := &ListReply{
		Invoices: list,
	}

	if config.Verbose {
		log.Printf("invoice.List count=%d", len(list))
	}
	if e := writer.Encode(w, r, res); e != nil {
		httputil.LogErr("invoice.List", e)
	}
}

// validateHourFilePath validates that an HourFile path belongs to the expected entity.
// Path security checks (traversal, absolute paths) are handled by db.pathFilter.
func validateHourFilePath(path, entity string) error {
	return db.ValidateEntityPath(path, entity)
}

// Text returns the raw hour file associated with an invoice.
func Text(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Text with id=%s", name)
	}

	paths := db.InvoiceSearchPaths(entity, year, bucket, name)

	u := new(Invoice)
	e := db.View(func(t *db.Txn) error {
		if e := t.OpenFirst(paths, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		path := u.Meta.HourFile
		if path == "" {
			http.Error(w, "Invoice has no hourfile", http.StatusNotFound)
			return nil
		}

		if err := validateHourFilePath(path, entity); err != nil {
			return fmt.Errorf("validate path: %w", err)
		}

		f, e := t.OpenRaw(path)
		if e != nil {
			return fmt.Errorf("open hourfile: %w", e)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("close: %s", err)
			}
		}()

		if _, e := io.Copy(w, f); e != nil {
			return fmt.Errorf("copy: %w", e)
		}
		return nil
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.Text", e)
	}
}

// Pdf generates and returns a PDF version of an invoice.
func Pdf(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Pdf with id=%s", name)
	}

	paths := db.InvoiceSearchPaths(entity, year, bucket, name)

	var f *gofpdf.Fpdf
	u := new(Invoice)
	e := db.View(func(t *db.Txn) error {
		if e := t.OpenFirst(paths, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		f = pdf(db.Path+entity, u)
		return nil
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.Pdf", e)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, u.Meta.Invoiceid))
	w.Header().Set("Content-Type", "application/pdf")
	if e := f.Output(w); e != nil {
		httputil.InternalError(w, "invoice.Pdf output", e)
	}
}

// XML generates and returns a UBL XML version of an invoice.
func XML(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.XML with id=%s", name)
	}

	paths := db.InvoiceSearchPaths(entity, year, bucket, name)

	var ubl *bytes.Buffer
	u := new(Invoice)
	e := db.View(func(t *db.Txn) error {
		if e := t.OpenFirst(paths, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		var err error
		ubl, err = UBL(u)
		if err != nil {
			return fmt.Errorf("generate ubl: %w", err)
		}
		return nil
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.XML", e)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.xml"`, u.Meta.Invoiceid))
	w.Header().Set("Content-Type", "application/xml")
	if _, e := w.Write(ubl.Bytes()); e != nil {
		httputil.InternalError(w, "invoice.XML write", e)
	}
}

/*func Credit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	args := r.URL.Query()
	bucket := args.Get("bucket")
	if bucket == "" {
		bucket = "invoices"
	}
	if !strings.HasPrefix(bucket, "invoices") {
		http.Error(w, "invoice.Load invalid bucket-name", http.StatusBadRequest)
		return
	}

	log.Printf("invoice.Credit with id=%s", name)
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get([]byte(name))
		if v == nil {
			return fmt.Errorf("No such invoice name")
		}

		u := new(Invoice)
		if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
			return e
		}

		u.Meta.Issuedate = time.Now().Format("2006-01-02")
		u.Meta.Duedate = ""
		u.Meta.Conceptid = fmt.Sprintf("CREDIT-%d", randStringBytesRmndr(6))

		b = tx.Bucket([]byte("invoices"))
		buf := new(bytes.Buffer)
		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}
		return b.Put([]byte(u.Meta.Conceptid), buf.Bytes())

		w.Header().Set("Content-Type", "application/json")
		if e := json.NewEncoder(w).Encode(u); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Pdf fail", http.StatusInternalServerError)
	}
}*/
