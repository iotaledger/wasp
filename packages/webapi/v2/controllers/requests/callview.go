package requests

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/kv/dict"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
)

func (c *Controller) executeCallView(e echo.Context) error {
	var callViewRequest models.ContractCallViewRequest

	if err := e.Bind(&callViewRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	chainID, err := isc.ChainIDFromString(callViewRequest.ChainID)
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	// Get contract and function. The request model supports HName and common string names. HNames are preferred.
	contractHName := callViewRequest.ContractHName
	functionHName := callViewRequest.FunctionHName

	if contractHName == 0 {
		contractHName = isc.Hn(callViewRequest.ContractName)
	}

	if functionHName == 0 {
		functionHName = isc.Hn(callViewRequest.FunctionName)
	}

	args, err := dict.FromJSONDict(callViewRequest.Arguments)
	if err != nil {
		return apierrors.InvalidPropertyError("arguments", err)
	}

	result, err := c.vmService.CallViewByChainID(chainID, contractHName, functionHName, args)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	return e.JSON(http.StatusOK, result)
}
