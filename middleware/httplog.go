package middleware

import (
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPLog is middleware that logs HTTP requests with timing.
func HTTPLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		meta := ""
		if r.Method != http.MethodGet {
			meta = r.Header.Get("Content-Type")
		}
		begin := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(begin).String()
		log.Printf("HTTP[%s] %s %s duration=%s %s\n", strconv.Quote(r.RemoteAddr), strconv.Quote(r.Method), strconv.Quote(r.URL.String()), duration, strconv.Quote(meta))
	})
}

// ClearSessionCookie sends an expired cookie to clear the session.
func ClearSessionCookie(w http.ResponseWriter) {
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

// HTTPAuth is middleware that validates session authentication.
func HTTPAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := &Sess{}
		cookie, e := r.Cookie("sess")
		if e != nil && e.Error() != "http: named cookie not present" {
			log.Printf("HTTPAuth %s", strconv.Quote(e.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte("Failed reading cookie")); err != nil {
				log.Printf("HTTPAuth write: %s", err)
			}
			return
		}

		sessionValid := false
		if cookie != nil && len(cookie.Value) > 0 {
			if e := decryptSession(cookie.Value, sess); e != nil {
				log.Printf("HTTPAuth decrypt error: %s", strconv.Quote(e.Error()))
			} else if isSessionExpired(sess) {
				log.Printf("HTTPAuth session expired for %s", sess.Email)
				ClearSessionCookie(w)
				sess = &Sess{} // Clear session data
			} else {
				sessionValid = true
			}
		}

		requireAuth := strings.HasPrefix(r.URL.Path, "/api/v1/")
		if requireAuth {
			// Validate Referer/Origin for state-changing requests (CSRF protection)
			if !validateReferer(r) {
				log.Printf("HTTPAuth CSRF check failed for %s %s", strconv.Quote(r.Method), strconv.Quote(r.URL.Path))
				w.WriteHeader(http.StatusForbidden)
				if _, err := w.Write([]byte("CSRF validation failed")); err != nil {
					log.Printf("HTTPAuth write: %s", err)
				}
				return
			}

			if !sessionValid || sess.Email == "" {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Auth missing")); err != nil {
					log.Printf("HTTPAuth write: %s", err)
				}
				return
			}
			user := UserByEmail(sess.Email)
			if user == nil {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("No such user")); err != nil {
					log.Printf("HTTPAuth write: %s", err)
				}
				return
			}

			// check if allowed to open URL
			segments := strings.Split(r.URL.Path, "/")
			if len(segments) >= 5 {
				ok, err := CompanyAllowed(segments[4], sess.Email)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := w.Write([]byte("Permission check fail")); err != nil {
						log.Printf("HTTPAuth write: %s", err)
					}
					return
				}
				if !ok {
					w.WriteHeader(http.StatusForbidden)
					if _, err := w.Write([]byte("Permission denied: You cannot view company " + html.EscapeString(segments[4]))); err != nil {
						log.Printf("HTTPAuth write: %s", err)
					}
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

// LocalOnly is middleware that restricts access to localhost.
func LocalOnly(next http.Handler) http.Handler {
	localIPv4 := "127.0.0.1"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: localIPv6 = "::1" ?
		if !strings.HasPrefix(r.RemoteAddr, localIPv4) {
			w.WriteHeader(http.StatusForbidden)
			if _, err := w.Write([]byte("Security exception")); err != nil {
				log.Printf("LocalOnly write: %s", err)
			}
			return
		}

		// Allow!
		next.ServeHTTP(w, r)
	})
}
