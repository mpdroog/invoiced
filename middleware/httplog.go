package middleware

import (
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

// clearSessionCookie sends an expired cookie to clear the session
func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sess",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// isStateChangingMethod returns true for methods that modify state
func isStateChangingMethod(method string) bool {
	return method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"
}

// validateReferer checks Referer header for state-changing requests (CSRF protection)
func validateReferer(r *http.Request) bool {
	// Only check state-changing methods
	if !isStateChangingMethod(r.Method) {
		return true
	}

	referer := r.Header.Get("Referer")
	origin := r.Header.Get("Origin")

	// At least one of Referer or Origin should be present for state-changing requests
	if referer == "" && origin == "" {
		return false
	}

	// Get expected host from request
	expectedHost := r.Host
	if expectedHost == "" {
		expectedHost = r.URL.Host
	}

	// Check Origin header if present
	if origin != "" {
		// Origin should match our host
		if !strings.Contains(origin, expectedHost) {
			return false
		}
	}

	// Check Referer header if present
	if referer != "" {
		// Referer should contain our host
		if !strings.Contains(referer, expectedHost) {
			return false
		}
	}

	return true
}

func HTTPAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := &Sess{}
		cookie, e := r.Cookie("sess")
		if e != nil && e.Error() != "http: named cookie not present" {
			log.Printf("HTTPAuth %s", e.Error())
			w.WriteHeader(500)
			w.Write([]byte("Failed reading cookie"))
			return
		}

		sessionValid := false
		if cookie != nil && len(cookie.Value) > 0 {
			if e := decryptSession(cookie.Value, sess); e != nil {
				log.Printf("HTTPAuth decrypt error: %s", e.Error())
				clearSessionCookie(w, r)
			} else if isSessionExpired(sess) {
				log.Printf("HTTPAuth session expired for %s", sess.Email)
				clearSessionCookie(w, r)
				sess = &Sess{} // Clear session data
			} else {
				sessionValid = true
			}
		}

		requireAuth := strings.HasPrefix(r.URL.Path, "/api/v1/")
		if requireAuth {
			// Validate Referer/Origin for state-changing requests (CSRF protection)
			if !validateReferer(r) {
				log.Printf("HTTPAuth CSRF check failed for %s %s", r.Method, r.URL.Path)
				w.WriteHeader(403)
				w.Write([]byte("CSRF validation failed"))
				return
			}

			if !sessionValid || sess.Email == "" {
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
