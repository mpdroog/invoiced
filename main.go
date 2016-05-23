package main

import (
	"github.com/mitchellh/go-homedir"
	"github.com/kardianos/osext"
	"fmt"
	"sync"
    "net/http"
    "github.com/julienschmidt/httprouter"
    "log"
    "github.com/toqueteos/webbrowser"
    //"github.com/boltdb/bolt"
    "github.com/mpdroog/invoiced/invoice"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"

    "github.com/mpdroog/invoiced/pdf"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Works!")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
    c := pdf.Content{
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
    }
    if e := res.OutputFileAndClose("hello.pdf"); e != nil {
        log.Fatal(e)
    }
    return

    var wg sync.WaitGroup

	folderPath, err := osext.ExecutableFolder()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(folderPath)

    home, e := homedir.Dir()
    if e != nil {
        log.Fatal(e)    	
    }

    /*db, err := bolt.Open(home + "/billing.db", 0600, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()*/
    db, err := sql.Open("sqlite3", home + "/billing.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    router := httprouter.New()
    router.GET("/", Index)
    router.GET("/api/invoice", invoice.List)
	router.ServeFiles("/static/*filepath", http.Dir(folderPath + "/static"))

	wg.Add(1)
	go func() {
	    e := http.ListenAndServe("localhost:9999", router)
	    wg.Done()
	    if e != nil {
	    	log.Fatal(e)
	    }
    }()
	if e := webbrowser.Open("http://localhost:9999"); e != nil {
		log.Fatal(e)
	}

	wg.Wait()
}