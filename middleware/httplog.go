package middleware

import (
	"log"
	"net/http"
	"time"
)

func HTTPLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		meta := ""
		if r.Method != "GET" {
			meta = r.Header.Get("Content-Type")
		}
		begin := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Now().Sub(begin).String()
		log.Printf("HTTP[%s] %s %s duration=%s %s\n", r.RemoteAddr, r.Method, r.URL.String(), duration, meta)
	})
}
