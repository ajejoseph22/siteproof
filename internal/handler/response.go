// Package handler contains HTTP request handlers.
package handler

import (
	"encoding/json"
	"net/http"
)

// Response is a generic API response.
type Response struct {
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse is an API error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details.
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginatedResponse is a paginated API response.
type PaginatedResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// PaginationMetadata contains pagination information.
type PaginationMetadata struct {
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	Total      int    `json:"total,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(Response{Data: data}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// WriteErrorWithDetails writes an error response with details.
func WriteErrorWithDetails(w http.ResponseWriter, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// WritePaginated writes a paginated response.
func WritePaginated(w http.ResponseWriter, data interface{}, limit, offset, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(PaginatedResponse{
		Data: data,
		Pagination: PaginationMetadata{
			Limit:   limit,
			Offset:  offset,
			Total:   total,
			HasMore: offset+limit < total,
		},
	})
}

// WriteCursorPaginated writes a cursor-paginated response.
func WriteCursorPaginated(w http.ResponseWriter, data interface{}, limit int, nextCursor string, hasMore bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(PaginatedResponse{
		Data: data,
		Pagination: PaginationMetadata{
			Limit:      limit,
			Offset:     0,
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	})
}

// DecodeJSON decodes a JSON request body.
func DecodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
