package dashboard

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

// same as https://github.com/labstack/echo/blob/151ed6b3f150163352985448b5630ab69de40aa5/echo.go#L347
// but renders HTML instead of json
func UseHTMLErrorHandler(e *echo.Echo) {
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			if he.Internal != nil {
				if herr, ok := he.Internal.(*echo.HTTPError); ok {
					he = herr
				}
			}
		} else {
			he = &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			}
		}

		code := he.Code
		message := fmt.Sprintf("%s", he.Message)
		if e.Debug {
			message = err.Error()
		}

		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				err = c.NoContent(code)
			} else {
				err = c.HTML(code, fmt.Sprintf(`<h1>%d %s</h1>`, code, message))
			}
			if err != nil {
				e.Logger.Error(err)
			}
		}
	}
}
