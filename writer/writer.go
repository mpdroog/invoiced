// Package writer abstracts the used encoding lib to
// send data back to the client.
package writer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"gopkg.in/vmihailenco/msgpack.v2"
)

const (
	contentTypeJSON    = "application/json"
	contentTypeMsgpack = "application/x-msgpack"
)

func getType(ctypes string, accepts []string) string {
	for _, ctype := range strings.Split(ctypes, ",") {
		ctype = strings.ToLower(strings.TrimSpace(ctype))
		for _, accept := range accepts {
			if ctype == accept {
				return accept
			}
		}
	}
	return ctypes
}

// Decode reads and decodes the request body based on Content-Type.
func Decode(r *http.Request, d interface{}) error {
	ctype := r.Header.Get("Content-Type")
	if idx := strings.Index(ctype, ";"); idx > -1 {
		ctype = ctype[:idx]
	}

	switch ctype {
	case contentTypeJSON:
		defer func() {
			if err := r.Body.Close(); err != nil {
				log.Printf("close: %s", err)
			}
		}()
		return json.NewDecoder(r.Body).Decode(d)
	case contentTypeMsgpack:
		defer func() {
			if err := r.Body.Close(); err != nil {
				log.Printf("close: %s", err)
			}
		}()
		return msgpack.NewDecoder(r.Body).Decode(d)
	default:
		return fmt.Errorf("invalid Content-Type=%s", ctype)
	}
}

// Encode writes the response body based on Accept header.
func Encode(w http.ResponseWriter, r *http.Request, d interface{}) error {
	accept := getType(r.Header.Get("Accept"), []string{contentTypeJSON, contentTypeMsgpack})
	if override := r.URL.Query().Get("accept"); override != "" {
		// Browser override
		accept = override
	}
	if accept == "" {
		// Default to json with error
		d = fmt.Sprintf("Invalid ?accept=%s", accept)
		accept = contentTypeJSON
	}

	var (
		b []byte
		e error
	)
	switch accept {
	case contentTypeJSON:
		b, e = json.Marshal(&d)
		if e != nil {
			return e
		}
	case contentTypeMsgpack:
		b, e = msgpack.Marshal(&d)
		if e != nil {
			return e
		}
	default:
		return fmt.Errorf("invalid accept=%s", accept)
	}

	w.Header().Set("Content-Type", accept)
	_, e = w.Write(b) //nolint:gosec // G705: encoded JSON/msgpack with proper Content-Type
	return e
}
