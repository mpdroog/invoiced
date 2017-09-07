package entities

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
	"github.com/mpdroog/invoiced/middleware"
	"github.com/mpdroog/invoiced/writer"
)

type DetailRes struct {
	User *middleware.User
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
		User: middleware.UserByEmail(r.Header.Get("X-User-Email")),
		Entity: middleware.CompanyByName(ps.ByName("entity")),
	}
	if e := writer.Encode(w, r, res); e != nil {
		log.Printf("entities.Details " + e.Error())
	}
}