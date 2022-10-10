package chain

import (
	"encoding/json"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/labstack/echo/v4"
)

func (c *Controller) executeCallView(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	payload := e.Request().Body
	if payload == http.NoBody {
		return apierrors.BodyIsEmptyError()
	}

	var callViewRequest models.ContractCallViewRequest
	if err := json.NewDecoder(payload).Decode(&callViewRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	// Get contract and function hName. HNames are preferred.
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

func (c *Controller) getContracts(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	contracts, err := c.chainService.GetContracts(chainID)
	if err != nil {
		return err
	}

	contractList := make([]models.ContractInfo, 0, len(contracts))

	for hName, contract := range contracts {
		contractInfo := models.ContractInfo{
			Description: contract.Description,
			HName:       hName,
			Name:        contract.Name,
			ProgramHash: contract.ProgramHash,
		}

		contractList = append(contractList, contractInfo)
	}

	return e.JSON(http.StatusOK, contractList)
}
