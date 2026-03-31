package invoice

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/jung-kurt/gofpdf"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/writer"
	"gopkg.in/gomail.v1"
	"gopkg.in/validator.v2"
)

// sanitizeEmailHeader removes CRLF sequences to prevent email header injection
func sanitizeEmailHeader(s string) string {
	// Remove carriage returns and line feeds to prevent header injection
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

// Job represents an email job with recipients and attachments.
type Job struct {
	To      []string
	Subject string
	Text    string
	Files   []string
}

// Email sends an invoice via email with PDF and XML attachments.
func Email(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("id")
	bucket := ps.ByName("bucket")
	entity := ps.ByName("entity")
	year := ps.ByName("year")

	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	m := new(InvoiceMail)
	if e := writer.Decode(r, m); e != nil {
		httputil.BadRequest(w, "invoice.Email decode", e)
		return
	}
	if e := validator.Validate(m); e != nil {
		http.Error(w, fmt.Sprintf("invoice.Email failed validate=%s", e), http.StatusBadRequest)
		return
	}
	if config.Verbose {
		log.Printf("invoice.Email with id=%s", name)
	}

	paths := db.InvoiceSearchPaths(entity, year, bucket, name)

	var f *gofpdf.Fpdf
	u := new(Invoice)

	change := db.Commit{
		Name:    r.Header.Get("X-User-Name"),
		Email:   r.Header.Get("X-User-Email"),
		Message: fmt.Sprintf("Email invoice %s to %s", name, m.To),
	}
	e := db.Update(change, func(t *db.Txn) error {
		e := t.OpenFirst(paths, u)
		if e != nil {
			return e
		}
		f = pdf(db.Path+entity, u)

		buf := new(bytes.Buffer)
		if e := f.Output(buf); e != nil {
			return e
		}

		// Sanitize email headers to prevent CRLF injection
		sanitizedTo := make([]string, 0)
		for _, addr := range strings.Split(m.To, ",") {
			sanitizedTo = append(sanitizedTo, sanitizeEmailHeader(strings.TrimSpace(addr)))
		}
		sanitizedSubject := sanitizeEmailHeader(m.Subject)

		job := &Job{
			To:      sanitizedTo,
			Subject: sanitizedSubject,
			Text:    m.Body,
			Files: []string{
				paths[1], // TODO: assumption
			},
		}

		hourbuf := new(bytes.Buffer)
		if len(u.Meta.HourFile) > 0 {
			// Validate HourFile path to prevent path traversal attacks
			if err := validateHourFilePath(u.Meta.HourFile, entity); err != nil {
				return err
			}
			f, e := t.OpenRaw(u.Meta.HourFile)
			if e != nil {
				return e
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Printf("close: %s", err)
				}
			}()
			if _, e := io.Copy(hourbuf, f); e != nil {
				return e
			}
			job.Files = append(job.Files, u.Meta.HourFile)
		}

		ubl, e := UBL(u)
		if e != nil {
			return e
		}

		rnd := randStringBytesRmndr(8)
		if e := t.Save(fmt.Sprintf("%s/%s/%s/mails/%s.toml", entity, year, bucket, rnd), true, job); e != nil {
			return e
		}

		hostname, e := os.Hostname()
		if e != nil {
			return e
		}

		var conf config.Queue
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
		msg.SetHeader("Subject", conf.Subject+job.Subject)
		msg.SetBody("text/plain", job.Text)

		msg.Attach(&gomail.File{
			Name:     u.Meta.Invoiceid + ".pdf",
			MimeType: "application/pdf",
			Content:  buf.Bytes(),
		})
		msg.Attach(&gomail.File{
			Name:     u.Meta.Invoiceid + ".xml",
			MimeType: "application/xml",
			Content:  ubl.Bytes(),
		})
		if hourbuf.Len() > 0 {
			msg.Attach(&gomail.File{
				Name:     "hours.txt",
				MimeType: "plain/text",
				Content:  hourbuf.Bytes(),
			})
		}

		if config.Verbose {
			log.Printf("Email=%+v\n", msg)
		}

		mailer := gomail.NewMailer(conf.Host, conf.User, conf.Pass, conf.Port)
		return mailer.Send(msg)
	})
	if e != nil {
		httputil.BadRequest(w, "invoice.Email", e)
		return
	}
}
