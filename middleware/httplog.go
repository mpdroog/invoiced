package middleware

import (
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
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

// validateReferer checks Referer header for state-changing requests (CSRF protection).
// Returns nil if valid, or an error explaining the failure.
func validateReferer(r *http.Request) error {
	// Only check state-changing methods
	if !isStateChangingMethod(r.Method) {
		return nil
	}

	referer := r.Header.Get("Referer")
	origin := r.Header.Get("Origin")

	// At least one of Referer or Origin should be present for state-changing requests
	if referer == "" && origin == "" {
		return errors.New("missing both Origin and Referer headers")
	}

	// Get expected host from request
	// Check X-Forwarded-Host first (set by reverse proxies)
	expectedHost := r.Header.Get("X-Forwarded-Host")
	if expectedHost == "" {
		expectedHost = r.Host
	}
	if expectedHost == "" {
		expectedHost = r.URL.Host
	}

	// Check Origin header if present
	if origin != "" {
		originURL, err := url.Parse(origin)
		if err != nil {
			return fmt.Errorf("invalid Origin header %q: %w", origin, err)
		}
		if originURL.Host != expectedHost {
			return fmt.Errorf("origin host mismatch: got %q, expected %q", originURL.Host, expectedHost)
		}
	}

	// Check Referer header if present
	if referer != "" {
		refererURL, err := url.Parse(referer)
		if err != nil {
			return fmt.Errorf("invalid Referer header %q: %w", referer, err)
		}
		if refererURL.Host != expectedHost {
			return fmt.Errorf("referer host mismatch: got %q, expected %q", refererURL.Host, expectedHost)
		}
	}

	return nil
}

// HTTPAuth is middleware that validates session or API key authentication.
func HTTPAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuth := strings.HasPrefix(r.URL.Path, "/api/v1/")

		// Check for API key authentication first
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" && requireAuth {
			user, err := ValidateAPIKey(apiKey)
			if err != nil {
				log.Printf("HTTPAuth invalid API key from %s", strconv.Quote(r.RemoteAddr))
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Invalid API key")); err != nil {
					log.Printf("HTTPAuth write: %s", err)
				}
				return
			}

			// API key valid - check company access (skip CSRF for API clients)
			segments := strings.Split(r.URL.Path, "/")
			if len(segments) >= 5 {
				ok, err := CompanyAllowed(segments[4], user.Email)
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
			next.ServeHTTP(w, r)
			return
		}

		// Fall back to session-based authentication
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

		if requireAuth {
			// Validate Referer/Origin for state-changing requests (CSRF protection)
			if err := validateReferer(r); err != nil {
				log.Printf("HTTPAuth CSRF check failed for %s %s: %s", strconv.Quote(r.Method), strconv.Quote(r.URL.Path), err)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isLocal := strings.HasPrefix(r.RemoteAddr, "127.0.0.1") ||
			strings.HasPrefix(r.RemoteAddr, "[::1]")
		if !isLocal {
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
