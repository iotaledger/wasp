package requests

import (
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/labstack/echo/v4"
)

func (c *Controller) waitForRequestToFinish(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	requestID, err := isc.RequestIDFromString(e.Param("requestID"))
	if err != nil {
		return apierrors.InvalidPropertyError("requestID", err)
	}

	receipt, vmError, err := c.chainService.WaitForRequestProcessed(chainID, requestID, 30*time.Second)
	if err != nil {
		return err
	}

	mappedReceipt := models.MapReceiptResponse(receipt, vmError)

	return e.JSON(http.StatusOK, mappedReceipt)
}
