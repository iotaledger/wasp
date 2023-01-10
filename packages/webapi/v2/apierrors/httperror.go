package apierrors

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

type HTTPErrorResult struct {
	Message interface{}
	Error   error
}

type HTTPError struct {
	HTTPCode        int
	Message         interface{}
	AdditionalError error
}

func NewHTTPError(httpCode int, message interface{}, err error) *HTTPError {
	return &HTTPError{
		HTTPCode:        httpCode,
		Message:         message,
		AdditionalError: err,
	}
}

func HTTPErrorFromEchoError(httpError *echo.HTTPError) *HTTPError {
	return NewHTTPError(httpError.Code, httpError.Message, httpError.Internal)
}

func (he *HTTPError) Error() string {
	return fmt.Sprintf("error: %v", he.Message)
}

func (he *HTTPError) GetErrorResult() *HTTPErrorResult {
	return &HTTPErrorResult{
		Message: he.Message,
		Error:   he.AdditionalError,
	}
}
