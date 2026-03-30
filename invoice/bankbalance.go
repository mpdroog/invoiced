package invoice

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/invoice/camt053"
	"github.com/mpdroog/invoiced/writer"
)

// Reply contains the result of processing bank statements.
type Reply struct {
	OK  int
	ERR int
}

// Balance parses a bank balance in CAMT053 format and marks matching invoices as paid.
func Balance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	var buf bytes.Buffer
	file, header, e := r.FormFile("file")
	if e != nil {
		httputil.InternalError(w, "invoice.Balance FormFile", e)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}()

	name := strings.Split(header.Filename, ".")
	if strings.ToLower(name[1]) != "xml" {
		http.Error(w, "Sorry, not an XML-file", 400)
		return
	}

	if _, e := io.Copy(&buf, file); e != nil {
		httputil.InternalError(w, "invoice.Balance io.Copy", e)
		return
	}

	p, e := camt053.FilterPaymentsReceived(&buf)
	if e != nil {
		httputil.InternalError(w, "invoice.Balance FilterPaymentsReceived", e)
		return
	}

	res := new(Reply)
	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: "Read CAMT053 bankbalance",
	}
	e = db.Update(change, func(t *db.Txn) error {
		for _, payment := range p {
			if config.Verbose {
				log.Printf(
					"Parse payments(%s) %sEUR with comment=%s from=%s(%s)\n",
					payment.ID, payment.Amount, payment.Comment, payment.Name, payment.IBAN,
				)
			}

			ok, e := balanceSetPaid(t, entity, year, payment.Comment, payment.Date, payment.Amount)
			if e != nil {
				return e
			}

			if ok {
				res.OK++
			} else {
				res.ERR++
			}

			if !ok {
				log.Printf("Failed marking payment(%s) as paid\n", payment.Comment)
			} else if config.Verbose {
				log.Printf("Marked payment(%s) as paid\n", payment.Comment)
			}
		}
		return nil
	})
	if e != nil {
		httputil.InternalError(w, "invoice.Balance commit", e)
		return
	}

	if e := writer.Encode(w, r, res); e != nil {
		httputil.LogErr("invoice.Balance encode", e)
	}
}

// TODO: Something cleaner?
func balanceSetPaid(t *db.Txn, entity, year, name, payDate, amount string) (bool, error) {
	b := strings.Index(name, "Q")
	e := strings.Index(name, "-")
	bucket := "Q" + name[b+1:e]
	from := db.InvoicePath(entity, year, bucket, name, false)
	to := db.InvoicePath(entity, year, bucket, name, true)

	u := new(Invoice)
	if e := t.Open(from, u); e != nil {
		return false, e
	}
	if u.Total.Total != amount {
		log.Printf("WARN: Invoice(%s) amounts don't match %sEUR/%sEUR\n", name, u.Total.Total, amount)
		return false, nil
	}
	u.Meta.Paydate = payDate

	if e := t.Save(to, false, u); e != nil {
		return false, e
	}
	if e := t.Remove(from); e != nil {
		return false, e
	}
	return true, nil
}
