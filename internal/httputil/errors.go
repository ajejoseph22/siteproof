// Package httputil provides HTTP utility functions shared across packages.
package httputil

import (
	"errors"
	"net/http"

	"github.com/ajejoseph22/siteproof/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	CodeInternalError     = "INTERNAL_ERROR"
	CodeValidationError   = "VALIDATION_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeConflict          = "CONFLICT"
	CodeInvalidToken      = "INVALID_TOKEN"
	CodeTokenExpired      = "TOKEN_EXPIRED"
	CodeTicketLocked      = "TICKET_LOCKED"
	CodeInvalidTransition = "INVALID_TRANSITION"
)

// APIError represents an error mapped to an HTTP response.
type APIError struct {
	Status  int
	Code    string
	Message string
	Details interface{}
}

func (e *APIError) Error() string {
	return e.Message
}

// APIErrorFrom maps a domain or database error to an API error response.
func APIErrorFrom(err error) APIError {
	if err == nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Code:    CodeInternalError,
			Message: "internal server error",
		}
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return *apiErr
	}

	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		mapped := mapDomainError(domainErr.Err)
		if domainErr.Code != "" {
			mapped.Code = domainErr.Code
		}
		if domainErr.Message != "" {
			mapped.Message = domainErr.Message
		}
		return mapped
	}

	if validationErrs, ok := err.(domain.ValidationErrors); ok {
		return validationAPIError(validationErrs)
	}
	if validationErr, ok := err.(*domain.ValidationError); ok {
		return validationAPIError(domain.ValidationErrors{*validationErr})
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return APIError{
			Status:  http.StatusNotFound,
			Code:    CodeNotFound,
			Message: "resource not found",
		}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return mapPGError(pgErr)
	}

	if mapped, ok := mapDomainErrorIfMatch(err); ok {
		return mapped
	}

	return APIError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternalError,
		Message: "internal server error",
	}
}

// WriteErrorFrom writes a standardized error response for a given error.
func WriteErrorFrom(w http.ResponseWriter, err error) {
	apiErr := APIErrorFrom(err)
	WriteAPIError(w, apiErr)
}

// WriteAPIError writes a standardized error response.
func WriteAPIError(w http.ResponseWriter, apiErr APIError) {
	if apiErr.Details != nil {
		WriteErrorWithDetails(w, apiErr.Status, apiErr.Code, apiErr.Message, apiErr.Details)
		return
	}
	WriteError(w, apiErr.Status, apiErr.Code, apiErr.Message)
}

func validationAPIError(errs domain.ValidationErrors) APIError {
	details := make(map[string]string, len(errs))
	for _, err := range errs {
		field := err.Field
		if field == "" {
			field = "error"
		}
		if _, exists := details[field]; !exists {
			details[field] = err.Message
		}
	}

	return APIError{
		Status:  http.StatusBadRequest,
		Code:    CodeValidationError,
		Message: "validation failed",
		Details: details,
	}
}

func mapDomainError(err error) APIError {
	if mapped, ok := mapDomainErrorIfMatch(err); ok {
		return mapped
	}

	return APIError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternalError,
		Message: "internal server error",
	}
}

func mapDomainErrorIfMatch(err error) (APIError, bool) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return APIError{Status: http.StatusNotFound, Code: CodeNotFound, Message: "resource not found"}, true
	case errors.Is(err, domain.ErrUnauthorized):
		return APIError{Status: http.StatusUnauthorized, Code: CodeUnauthorized, Message: "unauthorized"}, true
	case errors.Is(err, domain.ErrForbidden):
		return APIError{Status: http.StatusForbidden, Code: CodeForbidden, Message: "forbidden"}, true
	case errors.Is(err, domain.ErrConflict):
		return APIError{Status: http.StatusConflict, Code: CodeConflict, Message: "resource conflict"}, true
	case errors.Is(err, domain.ErrValidation):
		return APIError{Status: http.StatusBadRequest, Code: CodeValidationError, Message: "validation failed"}, true
	case errors.Is(err, domain.ErrTicketLocked):
		return APIError{Status: http.StatusUnprocessableEntity, Code: CodeTicketLocked, Message: "ticket is locked and cannot be modified"}, true
	case errors.Is(err, domain.ErrInvalidTransition):
		return APIError{Status: http.StatusUnprocessableEntity, Code: CodeInvalidTransition, Message: "invalid status transition"}, true
	case errors.Is(err, domain.ErrInvalidToken):
		return APIError{Status: http.StatusUnauthorized, Code: CodeInvalidToken, Message: "invalid token"}, true
	case errors.Is(err, domain.ErrExpiredToken):
		return APIError{Status: http.StatusUnauthorized, Code: CodeTokenExpired, Message: "token has expired"}, true
	default:
		return APIError{}, false
	}
}

func mapPGError(pgErr *pgconn.PgError) APIError {
	switch pgErr.Code {
	case "23505":
		return APIError{Status: http.StatusConflict, Code: CodeConflict, Message: "resource conflict"}
	case "23503":
		return APIError{Status: http.StatusConflict, Code: CodeConflict, Message: "invalid reference"}
	case "23502", "23514", "22P02":
		return APIError{Status: http.StatusBadRequest, Code: CodeValidationError, Message: "validation failed"}
	default:
		return APIError{Status: http.StatusInternalServerError, Code: CodeInternalError, Message: "internal server error"}
	}
}
