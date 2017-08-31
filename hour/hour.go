package hour

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/validator.v2"
	"log"
	"net/http"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/config"
)

func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	//year := ps.ByName("year")
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to delete", 400)
		return
	}

	if e := db.Remove(fmt.Sprintf("%s/concepts/hours/%s.toml", entity, name)); e != nil {
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
	if e := json.NewDecoder(r.Body).Decode(u); e != nil {
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
	if e := db.Save(fmt.Sprintf("%s/concepts/hours/%s.toml", entity, u.Name), u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("hour.Save failed writing to disk"), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(u); e != nil {
		log.Printf(e.Error())
	}
}

func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	log.Printf("hour.Load with id=%s", name)

	entity := ps.ByName("entity")
	u := new(Hour)
	if e := db.Open(fmt.Sprintf("%s/concepts/hours/%s.toml", entity, name), u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("hour.Load failed loading file from disk"), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(u); e != nil {
		log.Printf(e.Error())
	}
}

func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// List(path string, p Pagination) ([]interface{}, error) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	dirs := []string{
		fmt.Sprintf("%s/%s/concepts/hours", entity, year),
		fmt.Sprintf("%s/%s/{all}/hours", entity, year),
	}

	var list []string
	mem := new(Hour)
	p, e := db.List(dirs, db.Pagination{From:0, Count:30}, mem, func(filename, file, path string) error {
		list = append(list, filename)
		return nil
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("hour.List failed scanning disk"), 400)
		return	
	}
	if config.Verbose {
		log.Printf("hour.List count=%d", len(list))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Pagination-Total", string(p.Total))
	if e := json.NewEncoder(w).Encode(list); e != nil {
		log.Printf(e.Error())
	}
}
