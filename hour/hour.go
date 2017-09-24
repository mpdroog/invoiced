package hour

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/validator.v2"
	"log"
	"net/http"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/writer"
	"github.com/mpdroog/invoiced/utils"
	"github.com/mpdroog/invoiced/invoice"
	"time"
)

func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to delete", 400)
		return
	}

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Delete concept hour %s", name),
	}
	e := db.Update(change, func(t *db.Txn) error {
		return t.Remove(fmt.Sprintf("%s/%s/concepts/hours/%s.toml", entity, year, name))
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "hour.Delete fail", http.StatusInternalServerError)
	}
}

func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	u := new(Hour)
	if e := writer.Decode(r, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Save failed decoding input", 400)
		return
	}
	if e := validator.Validate(u); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Save failed validate=%s", e), 400)
		return
	}

	entity := ps.ByName("entity")
	year := ps.ByName("year")

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Save concept hour %s", u.Name),
	}

	isNew := true
	if u.Status != "NEW" {
		isNew = false
	}

	e := db.Update(change, func(t *db.Txn) error {
		u.Status = "CONCEPT"
		return t.Save(fmt.Sprintf("%s/%s/concepts/hours/%s.toml", entity, year, u.Name), isNew, u)
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "hour.Delete fail", http.StatusInternalServerError)
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("hour.Save " + e.Error())
	}
}

func Bill(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	name := ps.ByName("id")

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Bill hours from %s", name),
	}

	u := new(Hour)
	path := fmt.Sprintf("%s/%s/concepts/hours/%s.toml", entity, year, name)
	bucketTo := fmt.Sprintf("Q%d", utils.YearQuarter(time.Now()))
	pathTo := fmt.Sprintf("%s/%s/%s/hours/%s.toml", entity, year, bucketTo, name)

	e := db.Update(change, func(t *db.Txn) error {
		// Move from concept to finalized quarter
		if e := t.Open(path, u); e != nil {
			return e
		}
		u.Status = "FINAL"
		if e := t.Save(pathTo, true, u); e != nil {
			return e
		}
		if e := t.Remove(path); e != nil {
			return e
		}
		// Next create concept invoice
		return invoice.HourToInvoice(entity, year, u.Project, name, u.Total, change.Email, t)
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "hour.Bill fail", http.StatusInternalServerError)
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("hour.Bill " + e.Error())
	}
}

func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	log.Printf("hour.Load with id=%s", name)

	entity := ps.ByName("entity")
	year := ps.ByName("year")
	bucket := ps.ByName("bucket")

	u := new(Hour)
	e := db.View(func(t *db.Txn) error {
		return t.Open(fmt.Sprintf("%s/%s/%s/hours/%s.toml", entity, year, bucket, name), u)
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("hour.Load failed loading file from disk"), 400)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("entities.Load " + e.Error())
	}
}

func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	dirs := []string{
		fmt.Sprintf("%s/%s/concepts/hours", entity, year),
		fmt.Sprintf("%s/%s/{all}/hours", entity, year),
	}

	var (
		e error
	)
	mem := new(Hour)
	list := make(map[string][]string)

	e = db.View(func(t *db.Txn) error {
		_, e := t.List(dirs, db.Pagination{From:0, Count:30}, mem, func(filename, file, fpath string) error {
			k := utils.BucketDir(fpath)
			list[k] = append(list[k], filename)
			return nil
		})
		return e
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("hour.List failed scanning disk"), 400)
		return	
	}
	if config.Verbose {
		log.Printf("hour.List count=%d", len(list))
	}
	if e := writer.Encode(w, r, list); e != nil {
		log.Printf("hour.List " + e.Error())
	}
}
