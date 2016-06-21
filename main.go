package main

import (
	"github.com/mitchellh/go-homedir"
	"github.com/kardianos/osext"
	"sync"
    "net/http"
    "github.com/julienschmidt/httprouter"
    "log"
    //"github.com/toqueteos/webbrowser"
    "github.com/boltdb/bolt"
    //isql "github.com/mpdroog/invoiced/sql"
    //"database/sql"
    //_ "github.com/mattn/go-sqlite3"

    "github.com/mpdroog/invoiced/invoice"
)

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    http.Redirect(w, r, r.URL.String() + "/static/", 301)
}

func main() {
    // TODO: Save invoices into out/invoices/YYYY/QN

    /*c := pdf.Content{
        CompanyName: "RootDev",
        From: []string{
            "M.P. Droog",
            "Dorpsstraat 236a",
            "Obdam, 1713HP, NL",
        },
        To: []string{
            "XS News B.V.",
            "New Yorkstraat 9-13",
            "1175 RD Lijnden",
        },
        Meta: map[string]string{
            "Invoice ID": "2016Q3-0001",
            "Issue Date": "2016-05-23",
            "PO Number": "-",
            "Due Date": "2016-05-31",
        },
        Lines: []pdf.Line{
            pdf.Line{"PPF", "50,00", "42,50", "2.125,00"},
        },
        TotalEx: "2.125,00",
        TotalTax: "446,25",
        TotalInc: "2.571,25",

        Notes: []string{"Please pay invoice before 2016-05-31...."},
        Banking: map[string]string{
            "IBAN": "NL1234",
            "VAT": "TAXNR",
            "CoC": "COCNR",
        },
    }
    res, e := pdf.Create(c)
    if e != nil {
        log.Fatal(e)        
    }*/
    /*if e := res.OutputFileAndClose("hello.pdf"); e != nil {
        log.Fatal(e)
    }
    return*/

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

    log.Printf("BoltDB=%s\n", home + "/billing.db")
    db, e := bolt.Open(home + "/billing.db", 0600, nil)
    if e != nil {
        log.Fatal(e)
    }
    defer db.Close()

    if e := invoice.Init(db); e != nil {
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
    /*router.GET("/api/invoicelines", invoicelines.Pending)
    router.DELETE("/api/invoicelines/:id", invoicelines.Delete)
    router.POST("/api/invoicelines", invoicelines.Add)
    router.PUT("/api/invoicelines/:id", invoicelines.Update)*/

    /*router.GET("/api/invoices", invoice.List)
    router.POST("/api/invoices", invoice.Create)
    router.GET("/api/invoices/:id", invoice.Get)
    router.POST("/api/invoice/:id/credit", invoice.Credit)*/

    //router.GET("/api/sql/all", isql.GetAll)
    //router.GET("/api/sql/row", isql.GetRow)

    router.GET("/api/invoices", invoice.List)
    router.POST("/api/invoice", invoice.Save)
    router.GET("/api/invoice/:id", invoice.Load)
    router.GET("/api/invoice/:id/pdf", invoice.Pdf)

	router.ServeFiles("/static/*filepath", http.Dir(folderPath + "/static"))

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