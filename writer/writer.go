package writer

import (
	"log"
	"net/http"
	"encoding/json"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func Encode(w http.ResponseWriter, r *http.Request, d interface{}) error {
	accept := r.Header.Get("Accept")
	if override := r.URL.Query().Get("accept"); override != "" {
		// Browser override
		accept = override
	}

	if accept == "application/json" {
		str, e := json.Marshal(&d)
		if e != nil {
			log.Fatal(e)
		}
		w.Header().Set("Content-Type", accept)
		w.Write([]byte(str))
		return nil
	}
	if accept == "application/x-msgpack" {
		b, e := msgpack.Marshal(&d)
		if e != nil {
			log.Fatal(e)
		}
		w.Header().Set("Content-Type", accept)
		w.Write(b)
		return nil
	}

	log.Fatal("Invalid accept=" + accept)
	return nil
}