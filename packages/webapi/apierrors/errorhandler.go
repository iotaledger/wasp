package apierrors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type GenericError struct {
	Error string
}

// HTTPErrorHandler must be hooked to an echo server to render instances
// of HTTPError as JSON
func HTTPErrorHandler(err error, c echo.Context) error {
	if echoError, ok := err.(*echo.HTTPError); ok {
		mappedError := HTTPErrorFromEchoError(echoError)
		return c.JSON(mappedError.HTTPCode, mappedError.GetErrorResult())
	}

	if apiError, ok := err.(*HTTPError); ok {
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				return c.NoContent(apiError.HTTPCode)
			}
			return c.JSON(apiError.HTTPCode, apiError.GetErrorResult())
		}
	}

	c.Echo().DefaultHTTPErrorHandler(err, c)
	return nil
}
