package entities

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/middleware"
	"github.com/mpdroog/invoiced/writer"
	"log"
	"net/http"
	"os"
)

type DetailRes struct {
	User   *middleware.User
	Entity *middleware.Entity
}

// List company's the user can administrate
func List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c, e := r.Cookie("sess")
	if e != nil {
		panic("Should not get here?")
	}
	res, e := middleware.Companies(c.Value)
	if e != nil {
		log.Printf("List=%s\n", e.Error())
		http.Error(w, "Failed reading entities", 500)
		return
	}
	if e := writer.Encode(w, r, res); e != nil {
		log.Printf("entities.List " + e.Error())
	}
}

// List current company+user details
func Details(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	res := &DetailRes{
		User:   middleware.UserByEmail(r.Header.Get("X-User-Email")),
		Entity: middleware.CompanyByName(ps.ByName("entity")),
	}
	if e := writer.Encode(w, r, res); e != nil {
		log.Printf("entities.Details " + e.Error())
	}
}

// Open a new year for accounting.
func Open(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if len(entity) == 0 || len(year) == 0 {
		http.Error(w, "Missing entity/year argument(s)", 400)
		return
	}

	quarters := []string{"Q1", "Q2", "Q3", "Q4"}
	for _, q := range quarters {
		base := db.Path + entity + "/" + year + "/" + q
		if e := os.MkdirAll(base+"/sales-invoices-unpaid", os.ModePerm); e != nil {
			panic(e)
		}
		if e := os.MkdirAll(base+"/sales-invoices-paid", os.ModePerm); e != nil {
			panic(e)
		}
		if e := os.MkdirAll(base+"/hours", os.ModePerm); e != nil {
			panic(e)
		}
	}

	if e := os.MkdirAll(db.Path+entity+"/"+year+"/concepts/sales-invoices", os.ModePerm); e != nil {
		panic(e)
	}
	if e := os.MkdirAll(db.Path+entity+"/"+year+"/concepts/hours", os.ModePerm); e != nil {
		panic(e)
	}

	http.Error(w, "Created directories.", 200)
}
