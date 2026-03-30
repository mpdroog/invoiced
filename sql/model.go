// Package sql provides database query utilities for the SQL backend.
package sql

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/writer"
)

var db *sql.DB

// Init initializes the SQL database connection.
func Init(d *sql.DB) error {
	db = d
	return nil
}

func get(ctx context.Context, sql string, limit int) ([]map[string]*string, error) {
	var res []map[string]*string

	rows, e := db.QueryContext(ctx, sql) //nolint:gosec // G701: admin SQL query endpoint, protected by auth
	if e != nil {
		return nil, e
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	cols, e := rows.Columns()
	if e != nil {
		return nil, e
	}

	i := 0
	for rows.Next() {
		ptr := make([]interface{}, len(cols))
		ctx := make([]*string, len(cols))

		for i := range ptr {
			ptr[i] = &ctx[i]
		}

		if e := rows.Scan(ptr...); e != nil {
			return nil, e
		}

		m := make(map[string]*string, len(cols))
		for i := 0; i < len(cols); i++ {
			m[cols[i]] = ctx[i]
		}

		res = append(res, m)
		i++
		if limit > 0 && i == limit {
			// limit results
			break
		}
	}
	if e := rows.Err(); e != nil {
		return nil, e
	}

	return res, nil
}

// GetRow executes a SQL query and returns a single row.
func GetRow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sql := r.URL.Query().Get("sql")
	if sql == "" {
		http.Error(w, "missing sql-arg", http.StatusBadRequest)
		return
	}

	res, e := get(r.Context(), sql, 1)
	if e != nil {
		log.Printf("sql.GetRow: %s", e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	var rep map[string]*string
	if len(res) > 0 {
		rep = res[0]
	}

	if e := writer.Encode(w, r, rep); e != nil {
		log.Printf("sql.GetRow encode: %s", e)
	}
}

// GetAll executes a SQL query and returns all rows.
func GetAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sql := r.URL.Query().Get("sql")
	if sql == "" {
		http.Error(w, "missing sql-arg", http.StatusBadRequest)
		return
	}

	res, e := get(r.Context(), sql, 0)
	if e != nil {
		log.Printf("sql.GetAll: %s", e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	if e := writer.Encode(w, r, res); e != nil {
		log.Printf("sql.GetAll encode: %s", e)
	}
}
