package main

import (
	"flag"
	"github.com/julienschmidt/httprouter"
	"github.com/kardianos/osext"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/entities"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/metrics"
	"github.com/mpdroog/invoiced/middleware"
	"github.com/mpdroog/invoiced/rules"
	"github.com/mpdroog/invoiced/taxes"
	"log"
	"net/http"
	"sync"
	"time"
)

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	url := r.URL
	url.Path = "/static/index.html"
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Auth login
	if e := r.ParseForm(); e != nil {
		log.Printf("Login: %s\n", e.Error())
		http.Error(w, "Parse form-failed", 400)
		return
	}

	email := r.FormValue("email")
	pass := r.FormValue("pass")
	if len(email) == 0 || len(pass) == 0 {
		http.Error(w, "Missing POST email/pass", 400)
		return
	}

	sess, e := middleware.Login(email, pass)
	if e != nil {
		log.Printf("Login: %s\n", e.Error())
		http.Error(w, "Login: Auth failed", 400)
		return
	}
	if len(sess) == 0 {
		http.Error(w, "Login: Invalid user/pass", 400)
		return
	}

	// TODO: Somewhere general to share with logout?
	http.SetCookie(w, &http.Cookie{
		Name:     "sess",
		Value:    sess,
		Expires:  time.Now().Add(time.Hour * 24 * 365),
		HttpOnly: true,
		Domain:   r.URL.Host,
		//Secure: config.HTTPSOnly,
	})

	url := r.URL
	url.Path = "/static/index.html"
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func main() {
	var (
		wg   sync.WaitGroup
		path string
	)
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&config.DbPath, "d", "acct", "Path to database")
	flag.StringVar(&config.HTTPListen, "h", "localhost:9999", "HTTP listening port")
	flag.StringVar(&path, "c", "./config.toml", "Path to config")
	flag.Parse()

	var e error
	config.CurDir, e = osext.ExecutableFolder()
	if e != nil {
		log.Fatal(e)
	}
	if config.Verbose {
		log.Printf("Curdir=%s\n", config.CurDir)
		log.Printf("DB=%s\n", config.DbPath)
	}

	if e := config.Open(path); e != nil {
		log.Fatal(e)
	}

	//db.AlwaysLowercase = true
	if e := db.Init(config.DbPath); e != nil {
		panic(e)
	}
	if e := middleware.Init(); e != nil {
		panic(e)
	}

	/*if e := invoice.Init(db); e != nil {
		log.Fatal(e)
	}
	if e := hour.Init(db); e != nil {
		log.Fatal(e)
	}
	if e := metrics.Init(db); e != nil {
		log.Fatal(e)
	}*/
	if e := rules.Init(); e != nil {
		log.Fatal(e)
	}

	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/", Login)

	router.GET("/api/v1/entities", entities.List)
	router.GET("/api/v1/entities/:entity/logo", entities.Logo)
	router.GET("/api/v1/entities/:entity/details", entities.Details)
	router.GET("/api/v1/entities/:entity/open/:year", entities.Open)

	router.GET("/api/v1/debtors/:entity/search", entities.Search)
	router.GET("/api/v1/projects/:entity/search", entities.ProjectSearch)

	router.GET("/api/v1/metrics/:entity/:year", metrics.Dashboard)

	router.GET("/api/v1/invoices/:entity/:year", invoice.List)
	router.POST("/api/v1/invoice/:entity/:year", invoice.Save)
	router.GET("/api/v1/invoice/:entity/:year/:bucket/:id", invoice.Load)
	router.POST("/api/v1/invoice/:entity/:year/:bucket/:id/finalize", invoice.Finalize)
	router.POST("/api/v1/invoice/:entity/:year/:bucket/:id/reset", invoice.Reset)
	router.GET("/api/v1/invoice/:entity/:year/:bucket/:id/pdf", invoice.Pdf)
	router.GET("/api/v1/invoice/:entity/:year/:bucket/:id/text", invoice.Text)
	router.GET("/api/v1/invoice/:entity/:year/:bucket/:id/xml", invoice.Xml)
	//router.GET("/api/v1/invoice/:id/credit", invoice.Credit)
	router.POST("/api/v1/invoice/:entity/:year/:bucket/:id/paid", invoice.Paid)
	router.POST("/api/v1/invoice/:entity/:year/:bucket/:id/email", invoice.Email)
	router.POST("/api/v1/invoice-balance/:entity/:year", invoice.Balance)
	router.DELETE("/api/v1/invoice/:entity/:year/:bucket/:id", invoice.Delete)

	router.GET("/api/v1/hours/:entity/:year", hour.List)
	router.POST("/api/v1/hour/:entity/:year/:bucket", hour.Save)
	router.GET("/api/v1/hour/:entity/:year/:bucket/:id", hour.Load)
	router.POST("/api/v1/hour/:entity/:year/:bucket/:id/bill", hour.Bill)
	router.DELETE("/api/v1/hour/:entity/:year/:bucket/:id", hour.Delete)

	router.GET("/api/v1/summary/:entity/:year", taxes.Summary)
	router.POST("/api/v1/taxes/:entity/:year/:quarter", taxes.Tax)

	router.ServeFiles("/static/*filepath", http.Dir(config.CurDir+"/static"))

	wg.Add(1)
	go func() {
		var e error
		var router http.Handler = router
		if config.HTTPListen == "localhost:9999" {
			router = middleware.LocalOnly(router)
		}
		router = middleware.HTTPAuth(router)
		if config.Verbose {
			log.Printf("Listening on %s\n", config.HTTPListen)
			e = http.ListenAndServe(config.HTTPListen, middleware.HTTPLog(router))
		} else {
			e = http.ListenAndServe(config.HTTPListen, router)
		}
		wg.Done()
		if e != nil {
			log.Fatal(e)
		}
	}()
	wg.Wait()
}
