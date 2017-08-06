package middleware

import (
	"log"
	"net/http"
	"time"
	"strings"
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

func LocalOnly(next http.Handler) http.Handler {
	localIPv4 := "127.0.0.1"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: localIPv6 = "::1" ?
		if !strings.HasPrefix(r.RemoteAddr, localIPv4) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Security exception"))
			return
		}

		// Allow!
		next.ServeHTTP(w, r)
	})
}