package apierrors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler() func(error, echo.Context) {
	return func(err error, c echo.Context) {
		switch err := err.(type) {
		case *echo.HTTPError:
			mappedError := HTTPErrorFromEchoError(err)
			_ = c.JSON(mappedError.HTTPCode, mappedError.GetErrorResult())
			return

		case *HTTPError:
			if !c.Response().Committed {
				if c.Request().Method == http.MethodHead { // Issue #608
					_ = c.NoContent(err.HTTPCode)
					return
				}

				_ = c.JSON(err.HTTPCode, err.GetErrorResult())
				return
			}
			c.Echo().DefaultHTTPErrorHandler(err, c)

		default:
			c.Echo().DefaultHTTPErrorHandler(err, c)
		}
	}
}
