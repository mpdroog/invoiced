package entities

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/writer"
)

// Project represents a billing project configuration.
type Project struct {
	Name         string
	Debtor       string
	BillingEmail []string
	NoteAdd      string
	HourRate     float64
	DueDays      int
	PO           string
	Street1      string
}

// ProjectSearch searches for projects by name.
func ProjectSearch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	args := r.URL.Query()
	query := strings.ToLower(args.Get("query"))

	var projectList map[string]Project
	e := db.View(func(t *db.Txn) error {
		return t.Open(db.ProjectsPath(entity), &projectList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.ProjectSearch", e)
		return
	}

	var out []Project
	for name, project := range projectList {
		name = strings.ToLower(name)
		if strings.Contains(name, query) {
			project.Name = name // trick for autocomplete :x
			out = append(out, project)
		}
	}
	if e := writer.Encode(w, r, out); e != nil {
		httputil.LogErr("entities.ProjectSearch", e)
	}
}

// GetProject retrieves a project by name within a transaction.
func GetProject(t *db.Txn, entity, prj string) (*Project, error) {
	var projectList map[string]Project
	if e := t.Open(db.ProjectsPath(entity), &projectList); e != nil {
		return nil, e
	}

	for name, project := range projectList {
		if name == prj {
			return &project, nil
		}
	}

	return nil, nil
}
