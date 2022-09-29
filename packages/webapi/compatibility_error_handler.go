package webapi

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
)

/*
CompatibilityHTTPErrorHandler differentiates V1/V2 error types and uses their respective handler functions.
*/
func CompatibilityHTTPErrorHandler(err error, c echo.Context) {
	// Use V1 error handler, if error is a V1 error
	_, ok := err.(*httperrors.HTTPError)

	if ok {
		apierrors.HTTPErrorHandler(err, c)
		return
	}

	// Use V2 error handler otherwise. This is also a catch-all for any other error type.
	apierrors.HTTPErrorHandler(err, c)
}
