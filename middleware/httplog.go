package middleware

import (
	"github.com/itshosted/webutils/encrypt"
	//"github.com/mpdroog/invoiced/db"
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

func HTTPAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := &Sess{}
		cookie, e := r.Cookie("sess")
		if e != nil && e.Error() != "http: named cookie not present" {
			log.Printf("HTTPAuth " + e.Error())
			w.WriteHeader(500)
			w.Write([]byte("Failed reading cookie"))
			return
		}

		if cookie != nil && len(cookie.Value) > 0 {
			if e := encrypt.DecryptBase64("aes", entities.IV, cookie.Value, &sess); e != nil {
				log.Printf("HTTPAuth " + e.Error())
				w.WriteHeader(500)
				w.Write([]byte("Failed reading cookie"))
				return
			}
		}

		requireAuth := strings.HasPrefix(r.URL.Path, "/api/v1/")
		if sess.Email == "" && requireAuth {
			w.WriteHeader(401)
			w.Write([]byte("Auth missing"))
			return
		}
		// TODO: expiry?

		// Allow!
		next.ServeHTTP(w, r)
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