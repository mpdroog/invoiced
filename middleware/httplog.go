package middleware

import (
	"github.com/itshosted/webutils/encrypt"
	"log"
	"net/http"
	"strings"
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
				// TODO: Somewhere general with login setcookie...
				http.SetCookie(w, &http.Cookie{
					Name:     "sess",
					Value:    "",
					Expires:  time.Now().Add(-1 * time.Hour),
					HttpOnly: true,
					Domain:   r.URL.Host,
					//Secure: config.HTTPSOnly,
				})
			}
		}

		requireAuth := strings.HasPrefix(r.URL.Path, "/api/v1/")
		if requireAuth {
			if sess.Email == "" {
				w.WriteHeader(401)
				w.Write([]byte("Auth missing"))
				return
			}
			user := UserByEmail(sess.Email)
			if user == nil {
				w.WriteHeader(401)
				w.Write([]byte("No such user"))
				return
			}

			// check if allowed to open URL
			segments := strings.Split(r.URL.Path, "/")
			if len(segments) >= 5 {
				ok, err := CompanyAllowed(segments[4], sess.Email)
				if err != nil {
					w.WriteHeader(500)
					w.Write([]byte("Permission check fail"))
					return
				}
				if !ok {
					w.WriteHeader(403)
					w.Write([]byte("Permission denied: You cannot view company " + segments[4]))
					return
				}
			}

			r.Header.Add("X-User-Email", user.Email)
			r.Header.Add("X-User-Name", user.Name)
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
