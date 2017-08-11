package invoice

import (
	"fmt"
	"bytes"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"github.com/mpdroog/invoiced/invoice/camt053"
	"github.com/mpdroog/invoiced/config"
	"strings"
	"io"
	"github.com/mpdroog/invoiced/db"
)

// Parse bankbalance in CAMT053-format
func Balance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
   var buf bytes.Buffer
    file, header, e := r.FormFile("file")
    if e != nil {
    	log.Printf(e.Error())
		http.Error(w, "Failed reading file", 500)
		return
    }
    defer file.Close()
    name := strings.Split(header.Filename, ".")
    if name[1] != "xml" {
    	http.Error(w, "Sorry, not an XML-file", 400)
		return
    }

    io.Copy(&buf, file)
	p, e := camt053.FilterPaymentsReceived(&buf)
	if e != nil {
    	log.Printf(e.Error())
		http.Error(w, "Failed parsing file", 500)
		return
	}

	for _, payment := range p {
		if config.Verbose {
			log.Printf(
				"Parse payments(%s) %sEUR with comment=%s from=%s(%s)",
				payment.Id, payment.Amount, payment.Comment, payment.Name, payment.IBAN,
			)
		}

		ok, e := balanceSetPaid(payment.Comment, payment.Date, payment.Amount)
		if e != nil {
			log.Printf(e.Error())
			http.Error(w, "Failed marking payments as paid", 500)
			return
		}

		if !ok {
			log.Printf("Failed marking payment(%s) as paid", payment.Comment)
		} else if config.Verbose {
			log.Printf("Marked payment(%s) as paid", payment.Comment)
		}
	}
}

func balanceSetPaid(name string, payDate string, amount string) (bool, error) {
	bucket := "2017Q3" // TODO???
	entity := "rootdev"

	u := new(Invoice)
	if e := db.Open(fmt.Sprintf("%s/%s/sales-invoices-unpaid/%s.toml", entity, bucket, name), u); e != nil {
		return false, e
	}
	if u.Total.Total != amount {
		log.Printf("WARN: Invoice(%s) amounts don't match %sEUR/%sEUR", name, u.Total.Total, amount)
		return false, nil
	}
	u.Meta.Paydate = payDate

	if e := db.Remove(fmt.Sprintf("%s/%s/sales-invoices-unpaid/%s.toml", entity, bucket, name)); e != nil {
		return false, e
	}
	if e := db.Save(fmt.Sprintf("%s/%s/sales-invoices-paid/%s.toml", entity, bucket, name), u); e != nil {
		return false, e
	}
	if e := db.Commit(); e != nil {
		return false, e
	}

	return true, nil
}