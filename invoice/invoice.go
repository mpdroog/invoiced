package invoice

import (
	"encoding/json"
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
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
	name := strings.ToLower(ps.ByName("id"))
	if name == "" {
		http.Error(w, "Please supply a name to delete", 400)
		return
	}
	// TODO: read from cookie?
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if e := db.Remove(fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, name)); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Delete fail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, e := w.Write([]byte("{'ok': true}")); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Delete fail", http.StatusInternalServerError)
		return
	}
}

// Lock invoice for changes and set invoiceid
func Finalize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := strings.ToLower(ps.ByName("id"))
	if name == "" {
		http.Error(w, "Please supply a name to finalize", 400)
		return
	}
	// TODO: read from cookie?
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Finalize with conceptid=%s", name)
	}

	from := fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, name)
	u := new(Invoice)
	if e := db.Open(from, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Finalize failed loading file from disk"), 400)
		return
	}

	if len(u.Meta.Issuedate) == 0 {
		u.Meta.Issuedate = time.Now().Format("2006-01-02")
	}

	if u.Meta.Invoiceid == "" {
		// Create invoiceid
		// TODO?????
		var idx uint64 = 22
		/*idx, e := b.NextSequence()
		if e != nil {
			return e
		}*/

		u.Meta.Invoiceid = createInvoiceId(time.Now(), idx)
		if config.Verbose {
			log.Printf("invoice.Finalize create conceptId=%s invoiceId=%s", u.Meta.Conceptid, u.Meta.Invoiceid)
		}
	}
	u.Meta.Status = "FINAL"

	// TODO: Uniqueness check?

	now, e := time.Parse("2006-01-02", u.Meta.Issuedate)
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Finalize failed parsing Issuedate"), 400)
		return
	}
	to := fmt.Sprintf("%s/%s/Q%d/sales-invoices-unpaid/%s.toml", entity, year, yearQuarter(now), name)
	if e := db.Save(to, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Finalize failed writing to disk"), 400)
		return
	}

	if e := db.Remove(from); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Finalize fail", http.StatusInternalServerError)
		return
	}
	if e := db.Commit(); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Finalize fail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(u); e != nil {
		log.Printf(e.Error())
	}
}

func Reset(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := strings.ToLower(ps.ByName("id"))
	bucket := ps.ByName("bucket") // 2017Q3
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Reset with conceptid=%s", name)
	}

	from := fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name)
	u := new(Invoice)
	if e := db.Open(from, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Reset failed loading file from disk"), 400)
		return
	}
	to := fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, u.Meta.Conceptid)
	u.Meta.Status = "CONCEPT"
	if e := db.Save(to, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Reset failed writing to disk"), 400)
		return
	}
	if e := db.Remove(from); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Reset fail", http.StatusInternalServerError)
		return
	}
	if e := db.Commit(); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Reset fail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(u); e != nil {
		log.Printf(e.Error())
	}
}

// Mark invoice as paid
func Paid(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := strings.ToLower(ps.ByName("id"))
	bucket := ps.ByName("bucket") // 2017Q3
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	if config.Verbose {
		log.Printf("invoice.Paid with conceptid=%s", name)
	}

	from := fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name)
	to := fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name)
	u := new(Invoice)
	if e := db.Open(from, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Paid failed loading file from disk"), 400)
		return
	}
	u.Meta.Paydate = time.Now().Format("2006-01-02")
	if e := db.Save(to, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Paid failed writing to disk"), 400)
		return
	}
	if e := db.Remove(from); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Paid fail", http.StatusInternalServerError)
		return
	}
	if e := db.Commit(); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Paid fail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(u); e != nil {
		log.Printf(e.Error())
	}
}

func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// IF POST create InvoiceID = 2016Q3-0001
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	u := new(Invoice)
	if e := json.NewDecoder(r.Body).Decode(u); e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Save failed to decode input", 400)
		return
	}
	if e := validator.Validate(u); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Save failed validate=%s", e), 400)
		return
	}

	if u.Meta.Conceptid == "" {
		u.Meta.Conceptid = fmt.Sprintf("CONCEPT-%s", randStringBytesRmndr(6))
		log.Printf("invoice.Save create conceptId=%s", u.Meta.Conceptid)
	} else {
		log.Printf("invoice.Save update conceptId=%s", u.Meta.Conceptid)
	}
	u.Meta.Status = "CONCEPT"

	if e := db.Save(fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, u.Meta.Conceptid), u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Save failed writing to disk"), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(u); e != nil {
		log.Printf(e.Error())
	}
}

func Load(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//args := r.URL.Query()
	name := ps.ByName("id")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	bucket := ps.ByName("bucket")
	log.Printf("invoice.Load with conceptid=%s", name)

	paths := []string{
		fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name),
		fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name),
	}
	if (bucket == "concepts") {
		paths = []string{fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, name)}
	}

	u := new(Invoice)
	if e := db.OpenFirst(paths, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Reset failed loading file from disk"), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(u); e != nil {
		log.Printf(e.Error())
	}
}

func List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	args := r.URL.Query()
	//bucket := args.Get("bucket")

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

	list := make(map[string][]*Invoice)
	mem := new(Invoice)
	p, e := db.List(paths, db.Pagination{From:from, Count:count}, &mem, func(filename, filepath, path string) error {
		/*if _, exist := list[filepath]; exist {
			return fmt.Errorf("Duplicate filename " + filename)
		}*/
		list[path] = append(list[path], mem)
		mem = new(Invoice)
		return nil
	})
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.List failed scanning disk"), 400)
		return	
	}
	if config.Verbose {
		log.Printf("invoice.List count=%d", len(list))
	}

	w.Header().Set("X-Pagination-Total", string(p.Total))
	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(list); e != nil {
		log.Printf(e.Error())
		return
	}
}

func Pdf(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := strings.ToLower(ps.ByName("id"))
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")
	/*if bucket == "" {
		bucket = "invoices"
	}*/
	if config.Verbose {
		log.Printf("invoice.Pdf with id=%s", name)
	}

	paths := []string{
		fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name),
		fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name),
	}

	u := new(Invoice)
	if e := db.OpenFirst(paths, u); e != nil {
		log.Printf(e.Error())
		http.Error(w, fmt.Sprintf("invoice.Pdf failed loading file from disk"), 400)
		return
	}
	f, e := pdf(u)
	if e != nil {
		log.Printf(e.Error())
		http.Error(w, "invoice.Pdf fail", http.StatusInternalServerError)
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
