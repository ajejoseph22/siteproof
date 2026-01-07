// Package handler contains HTTP request handlers.
package handler

import (
	"net/http"

	"github.com/ajejoseph22/siteproof/internal/httputil"
)

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	httputil.WriteJSON(w, status, data)
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	httputil.WriteError(w, status, code, message)
}

// WriteErrorWithDetails writes an error response with details.
func WriteErrorWithDetails(w http.ResponseWriter, status int, code, message string, details interface{}) {
	httputil.WriteErrorWithDetails(w, status, code, message, details)
}

// WriteErrorFrom writes a standardized error response.
func WriteErrorFrom(w http.ResponseWriter, err error) {
	httputil.WriteErrorFrom(w, err)
}

// WritePaginated writes a paginated response.
func WritePaginated(w http.ResponseWriter, data interface{}, limit, offset, total int) {
	httputil.WritePaginated(w, data, limit, offset, total)
}

// WriteCursorPaginated writes a cursor-paginated response.
func WriteCursorPaginated(w http.ResponseWriter, data interface{}, limit int, nextCursor string, hasMore bool) {
	httputil.WriteCursorPaginated(w, data, limit, nextCursor, hasMore)
}

// DecodeJSON decodes a JSON request body.
func DecodeJSON(r *http.Request, v interface{}) error {
	return httputil.DecodeJSON(r, v)
}
