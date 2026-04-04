package entities

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/idx"
	"github.com/mpdroog/invoiced/writer"
)

// DebtorItem wraps a debtor with its key for API responses.
type DebtorItem struct {
	Key         string `json:"key"`
	Debtor      Debtor `json:"debtor"`
	LastInvoice string `json:"lastInvoice,omitempty"` // Last invoice date (YYYY-MM-DD)
}

// ProjectItem wraps a project with its key for API responses.
type ProjectItem struct {
	Key     string  `json:"key"`
	Project Project `json:"project"`
}

// DebtorList returns all debtors for an entity.
func DebtorList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))

	var debtorList map[string]Debtor
	e := db.View(func(t *db.Txn) error {
		return t.Open(db.DebtorsPath(entity), &debtorList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.DebtorList", e)
		return
	}

	// Get last invoice dates per customer name
	lastInvoices, err := idx.GetLastInvoiceDates(entity)
	if err != nil {
		httputil.LogErr("entities.DebtorList idx", err)
		// Continue without last invoice dates
		lastInvoices = make(map[string]string)
	}

	out := []DebtorItem{}
	for key, debtor := range debtorList {
		item := DebtorItem{Key: key, Debtor: debtor}
		// Match by debtor name (case-insensitive)
		if lastInvoice, ok := lastInvoices[strings.ToLower(debtor.Name)]; ok {
			item.LastInvoice = lastInvoice
		}
		out = append(out, item)
	}
	if e := writer.Encode(w, r, out); e != nil {
		httputil.LogErr("entities.DebtorList", e)
	}
}

// DebtorLoad returns a single debtor by key.
func DebtorLoad(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	key := ps.ByName("id")

	var debtorList map[string]Debtor
	e := db.View(func(t *db.Txn) error {
		return t.Open(db.DebtorsPath(entity), &debtorList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.DebtorLoad", e)
		return
	}

	debtor, ok := debtorList[key]
	if !ok {
		http.Error(w, "Debtor not found", http.StatusNotFound)
		return
	}

	if e := writer.Encode(w, r, DebtorItem{Key: key, Debtor: debtor}); e != nil {
		httputil.LogErr("entities.DebtorLoad", e)
	}
}

// DebtorSave creates or updates a debtor.
func DebtorSave(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	entity := strings.ToLower(ps.ByName("entity"))
	key := ps.ByName("id")

	var item DebtorItem
	if e := writer.Decode(r, &item); e != nil {
		httputil.BadRequest(w, "entities.DebtorSave decode", e)
		return
	}

	// Use URL key if provided, otherwise use key from body
	if key != "" {
		item.Key = key
	}
	if item.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	var action db.CommitAction
	e := db.View(func(t *db.Txn) error {
		var debtorList map[string]Debtor
		if err := t.Open(db.DebtorsPath(entity), &debtorList); err != nil {
			// File doesn't exist yet, so this is a create operation
			action = db.ActionCreate
			return nil //nolint:nilerr // intentionally ignoring "file not found" error
		}
		if _, exists := debtorList[item.Key]; exists {
			action = db.ActionUpdate
		} else {
			action = db.ActionCreate
		}
		return nil
	})
	if e != nil {
		httputil.InternalError(w, "entities.DebtorSave check", e)
		return
	}

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, action, db.ResourceDebtor, item.Key),
	}

	e = db.Update(&change, func(t *db.Txn) error {
		var debtorList map[string]Debtor
		if err := t.Open(db.DebtorsPath(entity), &debtorList); err != nil {
			// File might not exist yet
			debtorList = make(map[string]Debtor)
		}

		debtorList[item.Key] = item.Debtor
		return t.Save(db.DebtorsPath(entity), false, debtorList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.DebtorSave", e)
		return
	}

	if e := writer.Encode(w, r, item); e != nil {
		httputil.LogErr("entities.DebtorSave", e)
	}
}

// DebtorDelete removes a debtor by key.
func DebtorDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	key := ps.ByName("id")

	if key == "" {
		http.Error(w, "Please supply a key to delete", http.StatusBadRequest)
		return
	}

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, db.ActionDelete, db.ResourceDebtor, key),
	}

	e := db.Update(&change, func(t *db.Txn) error {
		var debtorList map[string]Debtor
		if err := t.Open(db.DebtorsPath(entity), &debtorList); err != nil {
			return err
		}

		if _, ok := debtorList[key]; !ok {
			return fmt.Errorf("debtor not found: %s", key)
		}

		delete(debtorList, key)
		return t.Save(db.DebtorsPath(entity), false, debtorList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.DebtorDelete", e)
		return
	}

	if e := writer.Encode(w, r, map[string]bool{"ok": true}); e != nil {
		httputil.LogErr("entities.DebtorDelete", e)
	}
}

// ProjectList returns all projects for an entity.
func ProjectList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))

	var projectList map[string]Project
	e := db.View(func(t *db.Txn) error {
		return t.Open(db.ProjectsPath(entity), &projectList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.ProjectList", e)
		return
	}

	out := []ProjectItem{}
	for key, project := range projectList {
		out = append(out, ProjectItem{Key: key, Project: project})
	}
	if e := writer.Encode(w, r, out); e != nil {
		httputil.LogErr("entities.ProjectList", e)
	}
}

// ProjectLoad returns a single project by key.
func ProjectLoad(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	key := ps.ByName("id")

	var projectList map[string]Project
	e := db.View(func(t *db.Txn) error {
		return t.Open(db.ProjectsPath(entity), &projectList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.ProjectLoad", e)
		return
	}

	project, ok := projectList[key]
	if !ok {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if e := writer.Encode(w, r, ProjectItem{Key: key, Project: project}); e != nil {
		httputil.LogErr("entities.ProjectLoad", e)
	}
}

// ProjectSave creates or updates a project.
func ProjectSave(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	entity := strings.ToLower(ps.ByName("entity"))
	key := ps.ByName("id")

	var item ProjectItem
	if e := writer.Decode(r, &item); e != nil {
		httputil.BadRequest(w, "entities.ProjectSave decode", e)
		return
	}

	// Use URL key if provided, otherwise use key from body
	if key != "" {
		item.Key = key
	}
	if item.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	var action db.CommitAction
	e := db.View(func(t *db.Txn) error {
		var projectList map[string]Project
		if err := t.Open(db.ProjectsPath(entity), &projectList); err != nil {
			// File doesn't exist yet, so this is a create operation
			action = db.ActionCreate
			return nil //nolint:nilerr // intentionally ignoring "file not found" error
		}
		if _, exists := projectList[item.Key]; exists {
			action = db.ActionUpdate
		} else {
			action = db.ActionCreate
		}
		return nil
	})
	if e != nil {
		httputil.InternalError(w, "entities.ProjectSave check", e)
		return
	}

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, action, db.ResourceProject, item.Key),
	}

	e = db.Update(&change, func(t *db.Txn) error {
		var projectList map[string]Project
		if err := t.Open(db.ProjectsPath(entity), &projectList); err != nil {
			// File might not exist yet
			projectList = make(map[string]Project)
		}

		projectList[item.Key] = item.Project
		return t.Save(db.ProjectsPath(entity), false, projectList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.ProjectSave", e)
		return
	}

	if e := writer.Encode(w, r, item); e != nil {
		httputil.LogErr("entities.ProjectSave", e)
	}
}

// ProjectDelete removes a project by key.
func ProjectDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))
	key := ps.ByName("id")

	if key == "" {
		http.Error(w, "Please supply a key to delete", http.StatusBadRequest)
		return
	}

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: db.FormatCommitMsg(entity, db.ActionDelete, db.ResourceProject, key),
	}

	e := db.Update(&change, func(t *db.Txn) error {
		var projectList map[string]Project
		if err := t.Open(db.ProjectsPath(entity), &projectList); err != nil {
			return err
		}

		if _, ok := projectList[key]; !ok {
			return fmt.Errorf("project not found: %s", key)
		}

		delete(projectList, key)
		return t.Save(db.ProjectsPath(entity), false, projectList)
	})
	if e != nil {
		httputil.InternalError(w, "entities.ProjectDelete", e)
		return
	}

	if e := writer.Encode(w, r, map[string]bool{"ok": true}); e != nil {
		httputil.LogErr("entities.ProjectDelete", e)
	}
}
