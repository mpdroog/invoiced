// Package httputil provides HTTP handler utilities for consistent
// error handling and logging across the application.
package httputil

import (
	"log"
	"net/http"
)

// Error logs an error and sends an HTTP error response.
// Use this for handler errors to ensure consistent logging format.
func Error(w http.ResponseWriter, context string, err error, code int) {
	log.Printf("%s: %s", context, err.Error())
	http.Error(w, context+" failed", code)
}

// LogErr logs an error with a consistent format.
// Use this when you need to log but not send an HTTP response.
func LogErr(context string, err error) {
	log.Printf("%s: %s", context, err.Error())
}

// BadRequest logs and sends a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, context string, err error) {
	Error(w, context, err, http.StatusBadRequest)
}

// InternalError logs and sends a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, context string, err error) {
	Error(w, context, err, http.StatusInternalServerError)
}

// NotFound logs and sends a 404 Not Found response.
func NotFound(w http.ResponseWriter, context string, err error) {
	Error(w, context, err, http.StatusNotFound)
}
