package httperrors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPErrorHandler must be hooked to an echo server to render instances
// of HTTPError as JSON
func HTTPErrorHandler(err error, c echo.Context) {
	he, ok := err.(*HTTPError)
	if ok {
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				err = c.NoContent(he.Code)
			} else {
				err = c.JSON(he.Code, he)
			}
		}
	}
	c.Echo().DefaultHTTPErrorHandler(err, c)
}
