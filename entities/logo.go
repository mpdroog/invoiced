package entities

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/httputil"
)

// Logo returns the company logo image.
func Logo(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))

	buffer := new(bytes.Buffer)
	e := db.View(func(t *db.Txn) error {
		fd, e := t.OpenRaw(db.LogoPath(entity))
		if e != nil {
			return e
		}
		defer func() {
			if err := fd.Close(); err != nil {
				log.Printf("entities.Logo close: %s", err)
			}
		}()

		buf := bufio.NewReader(fd)
		if _, e := io.Copy(buffer, buf); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		httputil.InternalError(w, "entities.Logo", e)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		httputil.LogErr("entities.Logo write", err)
	}
}
