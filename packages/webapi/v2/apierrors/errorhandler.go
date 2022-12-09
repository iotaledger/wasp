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
func HTTPErrorHandler(err error, c echo.Context) {
	he, ok := err.(*HTTPError)
	if ok {
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				err = c.NoContent(he.HTTPCode)
			} else {
				err = c.JSON(he.HTTPCode, he.GetErrorResult())
			}
		}
	} else {
		err = c.JSON(http.StatusInternalServerError, GenericError{Error: err.Error()})
	}

	c.Echo().DefaultHTTPErrorHandler(err, c)
}
