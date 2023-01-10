package httperrors

import (
	"net/http"
)

// HTTPError implements the Go error interface, and includes an HTTP response code.
// Any webapi endpoint can return an instance of HTTPError, and it will be rendered
// as JSON in the response.
type HTTPError struct {
	Code    int
	Message string
}

func (he *HTTPError) Error() string {
	return he.Message
}

func BadRequest(message string) *HTTPError {
	return &HTTPError{Code: http.StatusBadRequest, Message: message}
}

func NotFound(message string) *HTTPError {
	return &HTTPError{Code: http.StatusNotFound, Message: message}
}

func Conflict(message string) *HTTPError {
	return &HTTPError{Code: http.StatusConflict, Message: message}
}

func Timeout(message string) *HTTPError {
	return &HTTPError{Code: http.StatusRequestTimeout, Message: message}
}

func ServerError(message string) *HTTPError {
	return &HTTPError{Code: http.StatusInternalServerError, Message: message}
}
