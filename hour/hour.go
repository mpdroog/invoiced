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
)

func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	//bucket := ps.ByName("bucket")
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to delete", 400)
		return
	}

	change := db.Commit{
		Name: r.Header.Get("X-Name"),
		Email: r.Header.Get("X-Email"),
		Message: fmt.Sprintf("Delete invoice %s", name),
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
	/*if u.Name == "" {
		http.Error(w, "invoice.Save err, no Name given", http.StatusInternalServerError)
		return
	}*/
	if e := validator.Validate(u); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Save failed validate=%s", e), 400)
		return
	}

	entity := ps.ByName("entity")
	year := ps.ByName("year")

	change := db.Commit{
		Name: r.Header.Get("X-Name"),
		Email: r.Header.Get("X-Email"),
		Message: fmt.Sprintf("Save invoice %s", u.Name),
	}
	e := db.Update(change, func(t *db.Txn) error {
		return t.Save(fmt.Sprintf("%s/%s/concepts/hours/%s.toml", entity, year, u.Name), u)
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "hour.Delete fail", http.StatusInternalServerError)
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("hour.Save " + e.Error())
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
