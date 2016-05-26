package sql

import (
	"database/sql"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"github.com/mpdroog/invoiced/writer"
)

var db *sql.DB

func Init(d *sql.DB) error {
	db = d
	return nil
}

func get(sql string, max int) ([]map[string]*string, error) {
	var res []map[string]*string

	rows, e := db.Query(sql)
	if e != nil {
		log.Fatal(e)
	}
	defer rows.Close()
	cols, e := rows.Columns()
	if e != nil {
		log.Fatal(e)
	}

	i := 0
	for rows.Next() {
		ptr := make([]interface{}, len(cols))
		ctx := make([]*string, len(cols))

		for i := range ptr {
			ptr[i] = &ctx[i]
		}

		if e := rows.Scan(ptr...); e != nil {
			log.Fatal(e)
		}

		m := make(map[string]*string, len(cols))
		for i := 0; i < len(cols); i++ {
			m[cols[i]] = ctx[i]
		}

		res = append(res, m)
		i++
		if max > 0 && i == max {
			// limit results
			break
		}
	}
	if e := rows.Err(); e != nil {
		log.Fatal(e)
	}

	return res, e
}

func GetRow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sql := r.URL.Query().Get("sql")
	if (sql == "") {
		log.Fatal("Missing sql-arg")
	}

	res, e := get(sql, 1)
	if e != nil {
		log.Fatal(e)
	}

	var rep map[string]*string
	if len(res) > 0 {
		rep = res[0]
	}

	if e := writer.Encode(w, r, rep); e != nil {
		log.Fatal(e)
	}
}

func GetAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sql := r.URL.Query().Get("sql")
	if (sql == "") {
		log.Fatal("Missing sql-arg")
	}

	res, e := get(sql, 0)
	if e != nil {
		log.Fatal(e)
	}

	if e := writer.Encode(w, r, res); e != nil {
		log.Fatal(e)
	}
}