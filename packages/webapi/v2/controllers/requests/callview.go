package requests

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
	"github.com/labstack/echo/v4"
)

func (c *Controller) executeCallView(e echo.Context) error {
	var callViewRequest models.ContractCallViewRequest

	if err := e.Bind(callViewRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	chainID, err := isc.ChainIDFromString(callViewRequest.ChainID)
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	// Get contract and function. The request model supports HName and common string names. HNames are preferred.
	var contractHName = callViewRequest.ContractHName
	var functionHName = callViewRequest.FunctionHName

	if contractHName == 0 {
		contractHName, err = isc.HnameFromString(callViewRequest.ContractName)
		if err != nil {
			return apierrors.InvalidPropertyError("contractName", err)
		}
	}

	if functionHName == 0 {
		functionHName, err = isc.HnameFromString(callViewRequest.FunctionName)
		if err != nil {
			return apierrors.InvalidPropertyError("functionName", err)
		}
	}

	result, err := c.vmService.CallViewByChainID(chainID, contractHName, functionHName, callViewRequest.Arguments)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	return e.JSON(http.StatusOK, result)
}
