package requests

import (
	"encoding/base64"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
)

func (c *Controller) handleOffLedgerRequest(e echo.Context) error {
	request := new(models.OffLedgerRequestBody)
	if err := e.Bind(request); err != nil {
		return apierrors.InvalidOffLedgerRequestError(err)
	}

	chainID, err := isc.ChainIDFromString(request.ChainID)
	if err != nil {
		return apierrors.InvalidPropertyError("ChainID", err)
	}

	requestDecoded, err := base64.StdEncoding.DecodeString(request.Request)
	if err != nil {
		return apierrors.InvalidPropertyError("Request", err)
	}

	err = c.offLedgerService.EnqueueOffLedgerRequest(chainID, requestDecoded)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	return e.NoContent(http.StatusAccepted)
}
