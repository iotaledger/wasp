package httperrors

import "net/http"

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
