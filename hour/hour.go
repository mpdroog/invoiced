// Package hour provides API endpoints for hour registration management.
package hour

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/utils"
	"github.com/mpdroog/invoiced/writer"
	"gopkg.in/validator.v2"
)

// Delete removes a concept hour registration.
func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to delete", 400)
		return
	}

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Delete concept hour %s", name),
	}
	e := db.Update(change, func(t *db.Txn) error {
		return t.Remove(db.ConceptHourPath(entity, year, name))
	})
	if e != nil {
		httputil.InternalError(w, "hour.Delete", e)
	}
}

// Save creates or updates an hour registration.
func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	u := new(Hour)
	if e := writer.Decode(r, u); e != nil {
		httputil.BadRequest(w, "hour.Save decode", e)
		return
	}
	if e := validator.Validate(u); e != nil {
		http.Error(w, fmt.Sprintf("hour.Save failed validate=%s", e), 400)
		return
	}

	entity := ps.ByName("entity")
	year := ps.ByName("year")

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Save concept hour %s", u.Name),
	}

	isNew := u.Status == "NEW"

	e := db.Update(change, func(t *db.Txn) error {
		u.Status = "CONCEPT"
		return t.Save(db.ConceptHourPath(entity, year, u.Name), isNew, u)
	})
	if e != nil {
		httputil.InternalError(w, "hour.Save", e)
		return
	}

	// CLI forward to wrapper app
	fmt.Printf("cmd entity=%s year=%s hour=%s\n", entity, year, u.Name)

	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("hour.Save", e)
	}
}

// Bill converts hour registrations into an invoice.
func Bill(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	name := ps.ByName("id")

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Bill hours from %s", name),
	}

	invoiceID := ""
	u := new(Hour)
	path := db.ConceptHourPath(entity, year, name)
	bucketTo := fmt.Sprintf("Q%d", utils.YearQuarter(time.Now()))
	pathTo := db.HourPath(entity, year, bucketTo, name)

	e := db.Update(change, func(t *db.Txn) error {
		// Move from concept to finalized quarter
		if e := t.Open(path, u); e != nil {
			return fmt.Errorf("open: %w", e)
		}
		u.Status = "FINAL"
		if e := t.Save(pathTo, true, u); e != nil {
			return fmt.Errorf("save: %w", e)
		}
		if e := t.Remove(path); e != nil {
			return fmt.Errorf("remove: %w", e)
		}
		// Next create concept invoice
		var err error
		invoiceID, err = invoice.HourToInvoice(entity, year, u.Project, name, u.Total, change.Email, pathTo, r.Header.Get("X-User-Name"), t)
		if err != nil {
			return fmt.Errorf("hour to invoice: %w", err)
		}
		return nil
	})
	if e != nil {
		httputil.InternalError(w, "hour.Bill", e)
		return
	}

	w.Header().Set("X-Redirect-Invoice", invoiceID)
	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("hour.Bill", e)
	}
}

// Load retrieves a single hour registration by ID.
func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	bucket := ps.ByName("bucket")
	if config.Verbose {
		log.Printf("hour.Load with id=%s", name)
	}

	u := new(Hour)
	e := db.View(func(t *db.Txn) error {
		return t.Open(db.HourPath(entity, year, bucket, name), u)
	})
	if e != nil {
		httputil.BadRequest(w, "hour.Load", e)
		return
	}

	// CLI forward to wrapper app
	fmt.Printf("cmd entity=%s year=%s hour=%s\n", entity, year, u.Name)
	if e := writer.Encode(w, r, u); e != nil {
		httputil.LogErr("hour.Load", e)
	}
}

// List returns all hour registrations for a year.
func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	dirs := db.HourListPaths(entity, year)

	mem := new(Hour)
	list := make(map[string][]string)

	e := db.View(func(t *db.Txn) error {
		_, e := t.List(dirs, db.Pagination{From: 0, Count: 30}, mem, func(filename, _, fpath string) error {
			k := utils.BucketDir(fpath)
			list[k] = append(list[k], filename)
			return nil
		})
		return e
	})
	if e != nil {
		httputil.BadRequest(w, "hour.List", e)
		return
	}
	if config.Verbose {
		log.Printf("hour.List count=%d", len(list))
	}
	if e := writer.Encode(w, r, list); e != nil {
		httputil.LogErr("hour.List", e)
	}
}
