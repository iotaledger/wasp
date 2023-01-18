package apierrors

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

type HTTPErrorResult struct {
	Message interface{}
	Error   string
}

type HTTPError struct {
	HTTPCode        int
	Message         interface{}
	AdditionalError string
}

func NewHTTPError(httpCode int, message interface{}, err error) *HTTPError {
	httpError := &HTTPError{
		HTTPCode: httpCode,
		Message:  message,
	}

	if err != nil {
		httpError.AdditionalError = err.Error()
	}

	return httpError
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
