package entities

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/writer"
	"strings"
	"fmt"
	"log"
)

type Project struct {
	Name string
	Debtor string
	BillingEmail []string
	NoteAdd string
	HourRate string
}

func ProjectSearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	args := r.URL.Query()
	query := strings.ToLower(args.Get("query"))

	var projectList map[string]Project
	e := db.View(func(t *db.Txn) error {
		return t.Open(fmt.Sprintf("%s/projects.toml", entity), &projectList)
	})
	if e != nil {
		log.Printf("entities.ProjectSearch e=" + e.Error())
		http.Error(w, "Failed reading debtors", 500)
		return
	}

	var out []Project
	for name, project := range projectList {
		name = strings.ToLower(name)
		if strings.Contains(name, query) {
			project.Name = name // trick for autocomplete :x
			log.Printf("Contains %s/%s\n", query, name)
			out = append(out, project)
		}
	}
	if e := writer.Encode(w, r, out); e != nil {
		log.Printf("entities.ProjectSearch " + e.Error())
	}
}