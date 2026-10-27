package taxes

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
	"github.com/mpdroog/invoiced/model"
	"github.com/mpdroog/invoiced/utils"
	"github.com/mpdroog/invoiced/writer"
)

func day(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// Status reads {entity}/{year}/Qn/tax/ and detects which quarters are filed/paid.
func Status(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year, e := strconv.Atoi(ps.ByName("year"))
	if e != nil {
		httputil.BadRequest(w, "taxes.Status year", e)
		return
	}
	if strings.Contains(entity, "/") || strings.Contains(entity, "..") {
		httputil.BadRequest(w, "taxes.Status entity", errors.New("invalid entity"))
		return
	}

	today := day(time.Now())
	out := []model.QuarterStatus{}

	e = db.View(func(t *db.Txn) error {
		for q := 1; q <= 4; q++ {
			deadline := BTWDeadline(year, q)
			s := model.QuarterStatus{
				Quarter:   fmt.Sprintf("Q%d", q),
				PeriodEnd: utils.QuarterEnd(year, q).Format("2006-01-02"),
				Deadline:  deadline.Format("2006-01-02"),
				DaysLeft:  int(deadline.Sub(today).Hours() / 24),
			}

			base := fmt.Sprintf("%s%s/%d/Q%d", db.Path, entity, year, q)
			if _, e := os.Stat(base); e == nil {
				s.Active = true
			} else if !errors.Is(e, fs.ErrNotExist) {
				return e
			}

			files, e := t.RawList(base + "/tax")
			if e != nil && !errors.Is(e, fs.ErrNotExist) {
				return e
			}
			for _, f := range files {
				name := f.Name()
				switch {
				case filedRegex.MatchString(name):
					s.Filed = name
				case icpRegex.MatchString(name):
					s.ICP = name
				case paidRegex.MatchString(name):
					s.Paid = name
				}
			}
			out = append(out, s)
		}
		return nil
	})
	if e != nil {
		httputil.InternalError(w, "taxes.Status", e)
		return
	}

	if e := writer.Encode(w, r, out); e != nil {
		httputil.LogErr("taxes.Status encode", e)
	}
}
