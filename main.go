package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/kardianos/osext"
	"github.com/mitchellh/go-homedir"
	"log"
	"net/http"
	"sync"
	//"github.com/toqueteos/webbrowser"
	"github.com/boltdb/bolt"
	//isql "github.com/mpdroog/invoiced/sql"
	//"database/sql"
	//_ "github.com/mattn/go-sqlite3"

	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/invoice"
)

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.Redirect(w, r, r.URL.String()+"/static/", 301)
}

func main() {
	var wg sync.WaitGroup

	folderPath, e := osext.ExecutableFolder()
	if e != nil {
		log.Fatal(e)
	}
	log.Printf("Curdir=%s\n", folderPath)

	home, e := homedir.Dir()
	if e != nil {
		log.Fatal(e)
	}

	log.Printf("BoltDB=%s\n", home+"/billing.db")
	db, e := bolt.Open(home+"/billing.db", 0600, nil)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()

	if e := invoice.Init(db); e != nil {
		log.Fatal(e)
	}
	if e := hour.Init(db); e != nil {
		log.Fatal(e)
	}

	/*log.Println(home + "/billing.sqlite")
	  db, err := sql.Open("sqlite3", home + "/billing.sqlite")
	  if err != nil {
	      log.Fatal(err)
	  }
	  defer db.Close()*/

	/*if e := isql.Init(db); e != nil {
	    log.Fatal(err)
	}*/

	// TODO: Init db?

	router := httprouter.New()
	router.GET("/", Index)

	//router.GET("/api/sql/all", isql.GetAll)
	//router.GET("/api/sql/row", isql.GetRow)

	router.GET("/api/invoices", invoice.List)
	router.POST("/api/invoice", invoice.Save)
	router.GET("/api/invoice/:id", invoice.Load)
	router.GET("/api/invoice/:id/pdf", invoice.Pdf)
	router.GET("/api/invoice/:id/credit", invoice.Credit)

	router.GET("/api/hours", hour.List)
	router.POST("/api/hour", hour.Save)
	router.GET("/api/hour/:id", hour.Load)

	router.ServeFiles("/static/*filepath", http.Dir(folderPath+"/static"))

	wg.Add(1)
	go func() {
		httpListen := "localhost:9999"
		log.Printf("Listening on %s\n", httpListen)
		e := http.ListenAndServe(httpListen, router)
		wg.Done()
		if e != nil {
			log.Fatal(e)
		}
	}()
	//if e := webbrowser.Open("http://localhost:9999"); e != nil {
	//	log.Fatal(e)
	//}

	wg.Wait()
}
