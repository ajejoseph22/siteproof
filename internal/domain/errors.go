// Package domain contains core business logic and domain entities.
package domain

import (
	"errors"
	"fmt"
)

// Domain errors that can be returned by services.
var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("resource conflict")
	ErrValidation        = errors.New("validation error")
	ErrTicketLocked      = errors.New("ticket is locked and cannot be modified")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrExpiredToken      = errors.New("token has expired")
	ErrInvalidToken      = errors.New("invalid token")
)

// ValidationError represents a field-level validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", e[0].Error())
}

// Add appends a validation error.
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// DomainError wraps an error with additional context.
type DomainError struct {
	Err     error
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error.
func NewDomainError(err error, code, message string) *DomainError {
	return &DomainError{
		Err:     err,
		Code:    code,
		Message: message,
	}
}
