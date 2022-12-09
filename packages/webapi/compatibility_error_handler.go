package webapi

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
)

/*
CompatibilityHTTPErrorHandler differentiates V1/V2 error types and uses their respective handler functions.
*/
func CompatibilityHTTPErrorHandler(logger *logger.Logger) func(error, echo.Context) {
	return func(err error, c echo.Context) {
		logger.Errorf("Compatability Error Handler: %v", err)
		// Use V1 error handler, if error is a V1 error
		_, ok := err.(*httperrors.HTTPError)

		if ok {
			apierrors.HTTPErrorHandler(err, c)
			return
		}

		// Use V2 error handler otherwise. This is also a catch-all for any other error type.
		apierrors.HTTPErrorHandler(err, c)
	}
}
