package entities

import (
	"github.com/mpdroog/invoiced/db"
	"strconv"
	"strings"
	"bytes"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"fmt"
	"io"
	"bufio"
)

func Logo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := strings.ToLower(ps.ByName("entity"))

	buffer := new(bytes.Buffer)
	e := db.View(func(t *db.Txn) error {
		fd, e := t.OpenRaw(fmt.Sprintf("%s/logo.png", entity))
		if e != nil {
			return e
		}
		defer fd.Close()

		buf := bufio.NewReader(fd)
		if _, e := io.Copy(buffer, buf); e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Printf("Logo e=" + e.Error())
		http.Error(w, "Failed reading logo", 500)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
    if _, err := w.Write(buffer.Bytes()); err != nil {
        log.Println("unable to write image.")
    }
}
