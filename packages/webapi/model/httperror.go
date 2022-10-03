package model

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPError is the standard error response for all webapi endpoints, and also implements
// the Go error interface.
type HTTPError struct {
	// StatusCode is the associated HTTP status code (default: htttp.StatusInternalServerError)
	StatusCode int

	// Message is the error message
	Message string
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{Message: message, StatusCode: statusCode}
}

// Error implements the Go error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

// IsHTTPNotFound returns true if the error is an HTTPError with status code http.StatusNotFound
func IsHTTPNotFound(e error) bool {
	var he *HTTPError
	ok := errors.As(e, &he)
	return ok && he.StatusCode == http.StatusNotFound
}
