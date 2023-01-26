package requests

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getReceipt(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	requestID, err := params.DecodeRequestID(e)
	if err != nil {
		return err
	}

	receipt, vmError, err := c.vmService.GetReceipt(chainID, requestID)
	if err != nil {
		return apierrors.ReceiptError(err)
	}

	mappedReceipt := models.MapReceiptResponse(receipt, vmError)

	return e.JSON(http.StatusOK, mappedReceipt)
}
