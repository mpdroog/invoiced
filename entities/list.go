package entities

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"log"
	"github.com/mpdroog/invoiced/middleware"
)

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

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(res); e != nil {
		log.Printf("entities.List " + e.Error())
	}
	return

	/* TODO
	for _, user := range entities.User {
		if user.Email == sess.Email {
			// Found the user
			var companies []Entity
			for _, company := range user.Company {
				// TODO: Not crash if not found?
				companies = append(companies, entities.Company[company])
			}

			w.Header().Set("Content-Type", "application/json")
			if e := json.NewEncoder(w).Encode("Hoi"); e != nil {
				log.Printf("entities.List " + e.Error())
			}
			return
		}
	}
	w.WriteHeader(403)
	w.Write([]byte("Invalid session"))*/
}