package requests

import (
	"net/http"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

func (c *Controller) handleOffLedgerRequest(e echo.Context) error {
	controllerutils.SetOperation(e, "offledger")
	request := new(models.OffLedgerRequest)
	if err := e.Bind(request); err != nil {
		return apierrors.InvalidOffLedgerRequestError(err)
	}

	chainID, err := isc.ChainIDFromString(request.ChainID)
	if err != nil {
		return apierrors.InvalidPropertyError("ChainID", err)
	}

	// set chainID to be used by the prometheus metrics
	e.Set(controllerutils.EchoContextKeyChainID, chainID)

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	requestDecoded, err := iotago.DecodeHex(request.Request)
	if err != nil {
		return apierrors.InvalidPropertyError("Request", err)
	}

	err = c.offLedgerService.EnqueueOffLedgerRequest(chainID, requestDecoded)
	if err != nil {
		return apierrors.ContractExecutionError(err) // TODO contract execution error? doesn't seem right...
	}

	return e.NoContent(http.StatusAccepted)
}
