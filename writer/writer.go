package writer

import (
	"encoding/json"
	"gopkg.in/vmihailenco/msgpack.v2"
	"net/http"
	"fmt"
	"strings"
)

func getType(ctypes string, accepts []string) string {
	for _, ctype := range strings.Split(ctypes, ",") {
		for _, accept := range accepts {
			if strings.ToLower(strings.TrimSpace(ctype)) == accept {
				return accept
			}
		}
	}
	return ctypes
}

func Decode(r *http.Request, d interface{}) error {
	ctype := r.Header.Get("Content-Type")

	if ctype == "application/json" {
		defer r.Body.Close()
		return json.NewDecoder(r.Body).Decode(d);
	} else if ctype == "application/x-msgpack" {
		defer r.Body.Close()
		return msgpack.NewDecoder(r.Body).Decode(d);
	} else {
		return fmt.Errorf("Invalid Content-Type=%s", ctype)
	}
}

func Encode(w http.ResponseWriter, r *http.Request, d interface{}) error {
	accept := getType(r.Header.Get("Accept"), []string{"application/json", "application/x-msgpack"})
	if override := r.URL.Query().Get("accept"); override != "" {
		// Browser override
		accept = override
	}

	var (
		b []byte
		e error
	)
	if accept == "application/json" {
		str, e := json.Marshal(&d)
		if e != nil {
			return e
		}
		b = []byte(str)
	} else if accept == "application/x-msgpack" {
		b, e = msgpack.Marshal(&d)
		if e != nil {
			return e
		}
	} else {
		return fmt.Errorf("Invalid accept=" + accept)
	}

	w.Header().Set("Content-Type", accept)
	_, e = w.Write(b)
	return e
}
