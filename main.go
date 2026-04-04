// Package main is the entry point for the invoiced HTTP server.
package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/julienschmidt/httprouter"
	"github.com/kardianos/osext"
	"github.com/mpdroog/invoiced/api"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/entities"
	gitpkg "github.com/mpdroog/invoiced/git"
	"github.com/mpdroog/invoiced/hour"
	"github.com/mpdroog/invoiced/idx"
	"github.com/mpdroog/invoiced/invoice"
	"github.com/mpdroog/invoiced/metrics"
	"github.com/mpdroog/invoiced/middleware"
	"github.com/mpdroog/invoiced/purchase"
	"github.com/mpdroog/invoiced/rules"
	"github.com/mpdroog/invoiced/search"
	"github.com/mpdroog/invoiced/taxes"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	url := r.URL
	url.Path = "/static/index.html"
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Auth login - limit body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1024) // 1KB max for login form
	if e := r.ParseForm(); e != nil {
		log.Printf("Login: %s\n", strconv.Quote(e.Error()))
		http.Error(w, "Parse form-failed", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	pass := r.FormValue("pass")
	if len(email) == 0 || len(pass) == 0 {
		http.Error(w, "Missing POST email/pass", http.StatusBadRequest)
		return
	}

	sess, e := middleware.Login(email, pass)
	if e != nil {
		// Log specific error internally, return generic message to client
		switch {
		case errors.Is(e, middleware.ErrUserNotFound):
			log.Println("Login: user not found for", strconv.Quote(email))
			http.Error(w, "Login: Invalid user/pass", http.StatusBadRequest)
		case errors.Is(e, middleware.ErrInvalidPassword):
			log.Println("Login: invalid password for", strconv.Quote(email))
			http.Error(w, "Login: Invalid user/pass", http.StatusBadRequest)
		default:
			log.Println("Login error:", strconv.Quote(e.Error()))
			http.Error(w, "Login: Auth failed", http.StatusBadRequest)
		}
		return
	}

	// Set secure session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sess",
		Value:    sess,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 8), // 8 hours to match SessionMaxAge
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	url := r.URL
	url.Path = "/static/index.html"
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	middleware.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	// db.AlwaysLowercase = true
	if e := db.Init(config.DbPath); e != nil {
		panic(e)
	}

	// Initialize SQLite index
	if e := idx.Open(config.DbPath); e != nil {
		log.Fatal(e)
	}

	// Set up sync hook
	db.OnCommit = func(touchedPaths []string, movedPaths []struct{ From, To string }) {
		for _, m := range movedPaths {
			if err := idx.MovePath(config.DbPath, m.From, m.To); err != nil {
				log.Fatal(err)
			}
		}
		for _, p := range touchedPaths {
			if err := idx.SyncPath(config.DbPath, p); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Rebuild index if empty
	if idx.IsEmpty() {
		log.Printf("Index empty, rebuilding...")
		if err := idx.Rebuild(config.DbPath); err != nil {
			log.Fatal(err)
		}
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
	router.POST("/logout", Logout)
	router.GET("/api", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/api/v1", http.StatusTemporaryRedirect)
	})

	// API route registry (self-documenting)
	reg := api.NewRegistry()

	// Entities
	reg.GET("/api/v1/entities", "Entities", "List all entities/companies", entities.List)
	reg.GET("/api/v1/entities/:entity/logo", "Entities", "Get entity logo", entities.Logo)
	reg.GET("/api/v1/entities/:entity/details", "Entities", "Get entity details", entities.Details)
	reg.GET("/api/v1/entities/:entity/open/:year", "Entities", "Open entity for year", entities.Open)

	// Debtors
	reg.GET("/api/v1/debtors/:entity/search", "Debtors", "Search debtors by name", entities.Search)
	reg.GET("/api/v1/debtors/:entity", "Debtors", "List all debtors", entities.DebtorList)
	reg.GET("/api/v1/debtor/:entity/:id", "Debtors", "Get debtor by ID", entities.DebtorLoad)
	reg.POST("/api/v1/debtor/:entity/:id", "Debtors", "Update debtor", entities.DebtorSave)
	reg.POST("/api/v1/debtor/:entity", "Debtors", "Create debtor", entities.DebtorSave)
	reg.DELETE("/api/v1/debtor/:entity/:id", "Debtors", "Delete debtor", entities.DebtorDelete)

	// Projects
	reg.GET("/api/v1/projects/:entity/search", "Projects", "Search projects by name", entities.ProjectSearch)
	reg.GET("/api/v1/projects/:entity", "Projects", "List all projects", entities.ProjectList)
	reg.GET("/api/v1/project/:entity/:id", "Projects", "Get project by ID", entities.ProjectLoad)
	reg.POST("/api/v1/project/:entity/:id", "Projects", "Update project", entities.ProjectSave)
	reg.POST("/api/v1/project/:entity", "Projects", "Create project", entities.ProjectSave)
	reg.DELETE("/api/v1/project/:entity/:id", "Projects", "Delete project", entities.ProjectDelete)

	// Search
	reg.GET("/api/v1/search/:entity", "Search", "Full-text search invoices/hours", search.Search)

	// Metrics
	reg.GET("/api/v1/metrics/:entity/:year", "Metrics", "Get simple metrics", metrics.Dashboard)
	reg.GET("/api/v1/dashboard/:entity/:year", "Metrics", "Get full dashboard data", metrics.DashboardFull)

	// Invoices
	reg.GET("/api/v1/invoices/:entity/:year", "Invoices", "List invoices for year", invoice.List)
	reg.POST("/api/v1/invoice/:entity/:year", "Invoices", "Create invoice", invoice.Save)
	reg.GET("/api/v1/invoice/:entity/:year/:bucket/:id", "Invoices", "Get invoice by ID", invoice.Load)
	reg.POST("/api/v1/invoice/:entity/:year/:bucket/:id/finalize", "Invoices", "Finalize invoice (concept -> final)", invoice.Finalize)
	reg.POST("/api/v1/invoice/:entity/:year/:bucket/:id/reset", "Invoices", "Reset invoice to concept", invoice.Reset)
	reg.GET("/api/v1/invoice/:entity/:year/:bucket/:id/pdf", "Invoices", "Download invoice PDF", invoice.Pdf)
	reg.GET("/api/v1/invoice/:entity/:year/:bucket/:id/text", "Invoices", "Get invoice as plain text", invoice.Text)
	reg.GET("/api/v1/invoice/:entity/:year/:bucket/:id/xml", "Invoices", "Get invoice as UBL XML", invoice.XML)
	reg.POST("/api/v1/invoice/:entity/:year/:bucket/:id/paid", "Invoices", "Mark invoice as paid", invoice.Paid)
	reg.POST("/api/v1/invoice/:entity/:year/:bucket/:id/email", "Invoices", "Email invoice to customer", invoice.Email)
	reg.POST("/api/v1/invoice-balance/:entity/:year", "Invoices", "Import bank statement (CAMT053)", invoice.Balance)
	reg.DELETE("/api/v1/invoice/:entity/:year/:bucket/:id", "Invoices", "Delete invoice", invoice.Delete)

	// Purchases
	reg.GET("/api/v1/purchases/:entity/:year", "Purchases", "List purchase invoices", purchase.List)
	reg.POST("/api/v1/purchase/:entity/:year", "Purchases", "Upload purchase invoice (UBL)", purchase.Upload)
	reg.GET("/api/v1/purchase/:entity/:year/:bucket/:id", "Purchases", "Get purchase invoice", purchase.Load)
	reg.GET("/api/v1/purchase/:entity/:year/:bucket/:id/pdf", "Purchases", "Download purchase PDF", purchase.PDF)
	reg.POST("/api/v1/purchase/:entity/:year/:bucket/:id/paid", "Purchases", "Mark purchase as paid", purchase.Paid)
	reg.DELETE("/api/v1/purchase/:entity/:year/:bucket/:id", "Purchases", "Delete purchase invoice", purchase.Delete)

	// Hours
	reg.GET("/api/v1/hours/:entity/:year", "Hours", "List hour registrations", hour.List)
	reg.POST("/api/v1/hour/:entity/:year/:bucket", "Hours", "Create/update hours", hour.Save)
	reg.GET("/api/v1/hour/:entity/:year/:bucket/:id", "Hours", "Get hour registration", hour.Load)
	reg.POST("/api/v1/hour/:entity/:year/:bucket/:id/bill", "Hours", "Convert hours to invoice", hour.Bill)
	reg.DELETE("/api/v1/hour/:entity/:year/:bucket/:id", "Hours", "Delete hour registration", hour.Delete)

	// Taxes
	reg.GET("/api/v1/summary/:entity/:year", "Taxes", "Get tax summary (Excel export with ?excel=1)", taxes.Summary)
	reg.POST("/api/v1/taxes/:entity/:year/:quarter", "Taxes", "Get quarterly tax calculation", taxes.Tax)

	// Git
	reg.GET("/api/v1/git/:entity/status", "Git", "Get git status", gitpkg.Status)
	reg.GET("/api/v1/git/:entity/history", "Git", "Get commit history", gitpkg.History)
	reg.GET("/api/v1/git/:entity/diff/:hash", "Git", "Get diff for commit", gitpkg.Diff)
	reg.POST("/api/v1/git/:entity/push", "Git", "Push to remote", gitpkg.Push)
	reg.POST("/api/v1/git/:entity/pull", "Git", "Pull from remote", gitpkg.Pull)
	reg.POST("/api/v1/git/:entity/discard", "Git", "Discard all local changes", gitpkg.DiscardAll)
	reg.POST("/api/v1/git/:entity/reset/:hash", "Git", "Reset to specific commit", gitpkg.ResetTo)
	reg.POST("/api/v1/reindex", "Git", "Rebuild SQLite search index", gitpkg.RebuildIndex)

	// Register all API routes
	reg.Register(router)

	// API documentation endpoint
	router.GET("/api/v1", reg.DocsHandler())

	router.ServeFiles("/static/*filepath", http.Dir(config.CurDir+"/static"))

	wg.Add(1)
	go func() {
		var e error
		var router http.Handler = router
		if config.HTTPListen == "localhost:9999" {
			router = middleware.LocalOnly(router)
		}
		router = middleware.HTTPAuth(router)

		server := &http.Server{
			Addr:         config.HTTPListen,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		if config.Verbose {
			log.Printf("Listening on %s\n", config.HTTPListen)
			server.Handler = middleware.HTTPLog(router)
		} else {
			server.Handler = router
		}
		e = server.ListenAndServe()
		wg.Done()
		if e != nil {
			log.Fatal(e)
		}
	}()

	// Notify systemd that we're ready to accept connections
	sent, e := daemon.SdNotify(false, daemon.SdNotifyReady)
	if e != nil {
		log.Fatal(e)
	}
	if !sent && config.Verbose {
		log.Printf("SystemD notify NOT sent (not running under systemd)\n")
	}
	wg.Wait()
}
