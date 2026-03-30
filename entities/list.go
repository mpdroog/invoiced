package entities

import (
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/idx"
	"github.com/mpdroog/invoiced/middleware"
	"github.com/mpdroog/invoiced/writer"
)

// DetailRes contains user and entity details for the current session.
type DetailRes struct {
	User   *middleware.User
	Entity *middleware.Entity
}

// List returns companies the user can administrate.
func List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c, e := r.Cookie("sess")
	if e != nil {
		panic("Should not get here?")
	}
	res, e := middleware.Companies(c.Value)
	if e != nil {
		httputil.InternalError(w, "entities.List", e)
		return
	}

	// Collect years we have
	for entity, v := range res {
		base := db.Path + entity
		var years []string

		e = db.View(func(t *db.Txn) error {
			files, e := t.RawList(base)
			if e != nil {
				return e
			}
			for _, file := range files {
				if !file.IsDir() {
					continue
				}
				years = append(years, file.Name())
			}
			return nil
		})
		if e != nil {
			httputil.BadRequest(w, "entities.List scan", e)
			return
		}

		v.Years = years

		// Fetch revenue per year from index
		v.YearRevenue = make(map[string]string)
		for _, year := range years {
			yearInt, err := strconv.Atoi(year)
			if err != nil {
				continue
			}
			total, err := idx.GetYearlyTotal(entity, yearInt)
			if err != nil {
				httputil.LogErr("entities.List idx.GetYearlyTotal", err)
				continue
			}
			v.YearRevenue[year] = total
		}

		res[entity] = v
	}

	if e := writer.Encode(w, r, res); e != nil {
		httputil.LogErr("entities.List", e)
	}
}

// Details returns current company and user information.
func Details(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	res := &DetailRes{
		User:   middleware.UserByEmail(r.Header.Get("X-User-Email")),
		Entity: middleware.CompanyByName(ps.ByName("entity")),
	}
	if e := writer.Encode(w, r, res); e != nil {
		httputil.LogErr("entities.Details", e)
	}
}

// Open creates directories for a new accounting year.
func Open(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if len(entity) == 0 || len(year) == 0 {
		http.Error(w, "Missing entity/year argument(s)", 400)
		return
	}

	quarters := []string{"Q1", "Q2", "Q3", "Q4"}
	for _, q := range quarters {
		base := db.Path + entity + "/" + year + "/" + q
		if e := os.MkdirAll(base+"/sales-invoices-unpaid", 0750); e != nil {
			panic(e)
		}
		if e := os.MkdirAll(base+"/sales-invoices-paid", 0750); e != nil {
			panic(e)
		}
		if e := os.MkdirAll(base+"/hours", 0750); e != nil {
			panic(e)
		}
	}

	if e := os.MkdirAll(db.Path+entity+"/"+year+"/concepts/sales-invoices", 0750); e != nil {
		panic(e)
	}
	if e := os.MkdirAll(db.Path+entity+"/"+year+"/concepts/hours", 0750); e != nil {
		panic(e)
	}

	http.Error(w, "Created directories.", 200)
}
