package invoice

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/jung-kurt/gofpdf"
	//"encoding/base64"
	"bytes"
	"github.com/mpdroog/invoiced/writer"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"log"
	"gopkg.in/validator.v2"
	"fmt"
	"io"
	"strings"
	"gopkg.in/gomail.v1"
	"os"
)

type Job struct {
	To []string
	Subject string
	Text string
	Files []string
}

func Email(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	m := new(InvoiceMail)
	if e := writer.Decode(r, m); e != nil {
		log.Printf("invoice.Email " + e.Error())
		http.Error(w, "invoice.Email failed to decode input", 400)
		return
	}
	if e := validator.Validate(m); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Email failed validate=%s", e), 400)
		return
	}
	if config.Verbose {
		log.Printf("invoice.Email with id=%s", name)
	}

	paths := []string{
		fmt.Sprintf("%s/%s/%s/sales-invoices-paid/%s.toml", entity, year, bucket, name),
		fmt.Sprintf("%s/%s/%s/sales-invoices-unpaid/%s.toml", entity, year, bucket, name),
	}

	var f *gofpdf.Fpdf
	u := new(Invoice)

	change := db.Commit{
		Name: r.Header.Get("X-User-Name"),
		Email: r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Email invoice %s to %s", name, m.To),
	}
	e := db.Update(change, func(t *db.Txn) error {
		e := t.OpenFirst(paths, u)
		if e != nil {
			return e
		}
		f, e = pdf(u)
		if e != nil {
			return e
		}

		buf := new(bytes.Buffer)
		if e := f.Output(buf); e != nil {
			return e
		}

		job := &Job{
			To: strings.Split(m.To, ","),
			Subject: m.Subject,
			Text: m.Body,
			Files: []string{
				paths[1], // TODO: assumption
			},
		}

		hourbuf := new(bytes.Buffer)
		if len(u.Meta.HourFile) > 0 {
			f, e := t.OpenRaw(u.Meta.HourFile)
			if e != nil {
				return e
			}
			defer f.Close()
			if _, e := io.Copy(hourbuf, f); e != nil {
				return e
			}
			job.Files = append(job.Files, u.Meta.HourFile)
		}

		rnd := randStringBytesRmndr(8)
		if e := t.Save(fmt.Sprintf("%s/%s/%s/mails/%s.toml", entity, year, bucket, rnd), true, job); e != nil {
			return e
		}

		hostname, e := os.Hostname()
		if e != nil {
			return e
		}

		var conf config.ConfigQueue
		for _, conf = range config.C.Queues {
			// TODO: Nasty hack to get first item...
			break
		}

		msg := gomail.NewMessage()
		msg.SetHeader("Message-ID", fmt.Sprintf("<%s@%s>", rnd, hostname))
		msg.SetHeader("X-Mailer", "invoiced")
		msg.SetHeader("X-Priority", "3")

		msg.SetHeader("From", fmt.Sprintf("%s <%s>", conf.Display, conf.From))
		msg.SetHeader("Reply-To", fmt.Sprintf("%s <%s>", conf.Display, conf.FromReply))

		msg.SetHeader("To", job.To...)
		msg.SetHeader("Bcc", conf.BCC...)
		msg.SetHeader("Subject", conf.Subject + job.Subject)
		msg.SetBody("text/plain", job.Text)

		msg.Attach(&gomail.File{
			Name: u.Meta.Invoiceid + ".pdf",
			MimeType: "application/pdf",
			Content: buf.Bytes(),
		})
		if hourbuf.Len() > 0 {
			msg.Attach(&gomail.File{
				Name: "hours.txt",
				MimeType: "plain/text",
				Content: hourbuf.Bytes(),
			})
		}

		if config.Verbose {
			log.Printf("Email=%+v\n", msg)
		}

		mailer := gomail.NewMailer(conf.Host, conf.User, conf.Pass, conf.Port)
		return mailer.Send(msg)
	})
	if e != nil {
		log.Printf("invoice.Email " + e.Error())
		http.Error(w, fmt.Sprintf("invoice.Email failed loading file from disk"), 400)
		return
	}
}