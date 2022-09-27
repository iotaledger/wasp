package apierrors

import "fmt"

type HTTPErrorResult struct {
	Message string
	Error   error
}

type HTTPError struct {
	HTTPCode        int
	Message         string
	AdditionalError error
}

func NewHTTPError(httpCode int, message string, err error) *HTTPError {
	return &HTTPError{
		HTTPCode:        httpCode,
		Message:         message,
		AdditionalError: err,
	}
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
