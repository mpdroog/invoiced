package invoice

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/validator.v2"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/config"
	"github.com/jung-kurt/gofpdf"
	"github.com/mpdroog/invoiced/writer"
	"github.com/mpdroog/invoiced/utils"
	"io"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
const EUTaxComment = "VAT Reverse charge"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func randStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to delete", 400)
		return
	}
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Delete concept invoice %s", name),
	}
	e := db.Update(change, func(t *db.Txn) error {
		fmt.Printf(
			"%s/%s/concepts/sales-invoices/%s.toml", entity, year, name,
		)
		return t.Remove(fmt.Sprintf(
			"%s/%s/concepts/sales-invoices/%s.toml", entity, year, name,
		))
	})
	if e != nil {
		log.Printf("invoice.Delete " + e.Error())
		http.Error(w, "invoice.Delete fail", http.StatusInternalServerError)
		return		
	}

	// TODO: something cleaner?
	w.Header().Set("Content-Type", "application/json")
	if _, e := w.Write([]byte("{'ok': true}")); e != nil {
		log.Printf("invoice.Delete " + e.Error())
		http.Error(w, "invoice.Delete fail", http.StatusInternalServerError)
		return
	}
}

func vatCountry(nr string) string {
	country := "nl"
	if len(nr) > 3 {
		country = strings.ToLower(nr[0:2])
	}
	return country
}

// Lock invoice for changes and set invoiceid
func Finalize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	if name == "" {
		http.Error(w, "Please supply a name to finalize", 400)
		return
	}
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Finalize with conceptid=%s", name)
	}

	var u *Invoice
	bucketTo := ""
	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Finalize concept invoice %s", name),
	}
	e := db.Update(change, func(t *db.Txn) error {
		from := fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, name)
		u = new(Invoice)
		if e := t.Open(from, u); e != nil {
			return e
		}
		if len(u.Meta.Issuedate) == 0 {
			u.Meta.Issuedate = time.Now().Format("2006-01-02")
		}
		u.Meta.Status = "FINAL"

		if u.Meta.Invoiceid == "" {
			idx, e := NextInvoiceID(entity, t)
			if e != nil {
				return e
			}
			u.Meta.Invoiceid = utils.CreateInvoiceId(time.Now(), idx)
			if config.Verbose {
				log.Printf("invoice.Finalize create conceptId=%s invoiceId=%s", u.Meta.Conceptid, u.Meta.Invoiceid)
			}

			if vatCountry(u.Customer.Vat) != "nl" && !strings.Contains(u.Notes, EUTaxComment) {
				if len(u.Notes) > 0 {
					u.Notes += "\n\n"
				}
				u.Notes += EUTaxComment
			}
		}

		// TODO: Uniqueness check?

		now, e := time.Parse("2006-01-02", u.Meta.Issuedate)
		if e != nil {
			return e
		}
		bucketTo = fmt.Sprintf("Q%d", utils.YearQuarter(now))
		to := fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucketTo, name)
		if e := t.Save(to, true, u); e != nil {
			return e
		}
		if e := t.Remove(from); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf("invoice.Finalize " + e.Error())
		http.Error(w, "invoice.Finalize fail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Bucket-Change", bucketTo)
	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("invoice.Finalize " + e.Error())
	}
}

func Reset(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Reset with conceptid=%s", name)
	}

	from := fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name)
	bucketTo := "concepts"
	to := ""
	u := new(Invoice)

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Reset invoice to concept"),
	}
	e := db.Update(change, func(t *db.Txn) error {
		if e := t.Open(from, u); e != nil {
			return e
		}
		to = fmt.Sprintf("%s/%s/%s/sales-invoices/%s.toml", entity, year, bucketTo, u.Meta.Conceptid)
		u.Meta.Status = "CONCEPT"
		if e := t.Save(to, true, u); e != nil {
			return e
		}
		if e := t.Remove(from); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf("invoice.Reset " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.Reset failed loading file from disk"), 400)
		return		
	}

	w.Header().Set("X-Bucket-Change", bucketTo)
	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("invoice.Reset " + e.Error())
	}
}

// Mark invoice as paid
func Paid(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Paid with conceptid=%s", name)
	}

	from := fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name)
	to := fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name)
	u := new(Invoice)

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Mark invoice %s as paid", name),
	}
	e := db.Update(change, func(t *db.Txn) error {
		if e := t.Open(from, u); e != nil {
			return e
		}
		u.Meta.Paydate = time.Now().Format("2006-01-02")
		if e := t.Save(to, false, u); e != nil {
			return e
		}
		if e := t.Remove(from); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf("invoice.Paid " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.Paid failed loading file from disk"), 400)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("invoice.Paid " + e.Error())
	}
}

func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	u := new(Invoice)
	if e := writer.Decode(r, u); e != nil {
		log.Printf("invoice.Save " + e.Error())
		http.Error(w, "invoice.Save failed to decode input", 400)
		return
	}
	if e := validator.Validate(u); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Save failed validate=%s", e), 400)
		return
	}

	isNew := true
	if u.Meta.Status != "NEW" {
		isNew = false
	}

	if u.Meta.Conceptid == "" {
		u.Meta.Conceptid = fmt.Sprintf("concept-%s", randStringBytesRmndr(12))
		log.Printf("invoice.Save create conceptId=%s", u.Meta.Conceptid)
	} else {
		log.Printf("invoice.Save update conceptId=%s", u.Meta.Conceptid)
	}
	u.Meta.Status = "CONCEPT"

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Update invoice %s", u.Meta.Conceptid),
	}
	e := db.Update(change, func(t *db.Txn) error {
		return t.Save(fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, u.Meta.Conceptid), isNew, u)
	})
	if e != nil {
		log.Printf("invoice.Save " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.Save failed writing to disk"), 400)
		return
	}

	w.Header().Set("X-Bucket-Change", "concepts")
	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("invoice.Save " + e.Error())
	}
}

func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	bucket := ps.ByName("bucket")
	log.Printf("invoice.Load with conceptid=%s", name)

	var paths []string
	if (bucket == "concepts") {
		paths = []string{fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, name)}
	} else {
		paths = []string{
			fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name),
			fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name),
		}
	}

	u := new(Invoice)
	e := db.View(func(t *db.Txn) error {
		return t.OpenFirst(paths, u)
	})
	if e != nil {
		log.Printf("invoice.Load " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.Load failed loading file from disk"), 400)
		return
	}

	if e := writer.Encode(w, r, u); e != nil {
		log.Printf("invoice.Load " + e.Error())
	}
}

func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	args := r.URL.Query()

	paths := []string{
		fmt.Sprintf("%s/%s/concepts/sales-invoices", entity, year),
		fmt.Sprintf("%s/%s/{all}/sales-invoices-paid", entity, year),
		fmt.Sprintf("%s/%s/{all}/sales-invoices-unpaid", entity, year),
	}

	from, e := strconv.Atoi(args.Get("from"))
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.List fail", http.StatusInternalServerError)
		return
	}
	count, e := strconv.Atoi(args.Get("count"))
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.List fail", http.StatusInternalServerError)
		return
	}

	var inv []*db.CommitMessage
	list := make(map[string][]*Invoice)
	mem := new(Invoice)

	e = db.View(func(t *db.Txn) error {
		_, e := t.List(paths, db.Pagination{From:from, Count:count}, &mem, func(filename, filepath, fpath string) error {
			list[fpath] = append(list[fpath], mem)
			mem = new(Invoice)
			return nil
		})
		if e != nil {
			return e
		}
		e = t.Logs(3, func(c *db.CommitMessage) error {
			inv = append(inv, c)
			return nil
		})
		return e
	})
	if e != nil {
		log.Printf("invoice.List " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.List failed scanning disk"), 400)
		return
	}

	res := &ListReply{
		Invoices: list,
		Commits: inv,
	}

	if config.Verbose {
		log.Printf("invoice.List count=%d", len(list))
	}
	if e := writer.Encode(w, r, res); e != nil {
		log.Printf("invoice.List " + e.Error())
	}
}

func Text(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Text with id=%s", name)
	}

	paths := []string{
		fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name),
		fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name),
	}

	u := new(Invoice)
	e := db.View(func(t *db.Txn) error {
		e := t.OpenFirst(paths, u)
		if e != nil {
			return e
		}
		path := u.Meta.HourFile
		if path == "" {
			http.Error(w, fmt.Sprintf("Invoice has no hourfile"), 404)
			return nil
		}

		f, e := t.OpenRaw(path)
		if e != nil {
			return e
		}
		defer f.Close()

		if _, e := io.Copy(w, f); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf("invoice.Text " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.Text failed loading hourfile"), 400)
		return
	}
}

func Pdf(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Pdf with id=%s", name)
	}

	paths := []string{
		fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name),
		fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name),
	}

	var f *gofpdf.Fpdf
	u := new(Invoice)
	e := db.View(func(t *db.Txn) error {
		e := t.OpenFirst(paths, u)
		if e != nil {
			return e
		}
		f, e = pdf(u)
		return e
	})
	if e != nil {
		log.Printf("invoice.Pdf " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.Pdf failed loading file from disk"), 400)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, u.Meta.Invoiceid))
	w.Header().Set("Content-Type", "application/pdf")
	if e := f.Output(w); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Pdf fail", http.StatusInternalServerError)
		return
	}
}

/*func Credit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	args := r.URL.Query()
	bucket := args.Get("bucket")
	if bucket == "" {
		bucket = "invoices"
	}
	if !strings.HasPrefix(bucket, "invoices") {
		http.Error(w, "invoice.Load invalid bucket-name", 400)
		return
	}

	log.Printf("invoice.Credit with id=%s", name)
	e := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get([]byte(name))
		if v == nil {
			return fmt.Errorf("No such invoice name")
		}

		u := new(Invoice)
		if e := json.NewDecoder(bytes.NewBuffer(v)).Decode(u); e != nil {
			return e
		}

		u.Meta.Issuedate = time.Now().Format("2006-01-02")
		u.Meta.Duedate = ""
		u.Meta.Conceptid = fmt.Sprintf("CREDIT-%d", randStringBytesRmndr(6))

		b = tx.Bucket([]byte("invoices"))
		buf := new(bytes.Buffer)
		if e := json.NewEncoder(buf).Encode(u); e != nil {
			return e
		}
		return b.Put([]byte(u.Meta.Conceptid), buf.Bytes())

		w.Header().Set("Content-Type", "application/json")
		if e := json.NewEncoder(w).Encode(u); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Pdf fail", http.StatusInternalServerError)
	}
}*/
