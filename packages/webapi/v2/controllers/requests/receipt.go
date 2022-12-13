package requests

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/params"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
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

	receipt, vmError, err := c.vmService.GetReceipt(chainID, *requestID)
	if err != nil {
		return apierrors.ReceiptError(err)
	}

	mappedReceipt := models.MapReceiptResponse(receipt, vmError)

	return e.JSON(http.StatusOK, mappedReceipt)
}
