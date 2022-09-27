package chain

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"

	"github.com/labstack/echo/v4"
)

func (c *Controller) handleCallView(e echo.Context, chainID *isc.ChainID, contractHName, functionHName isc.Hname, payload io.ReadCloser) error {
	var params dict.Dict
	if payload != http.NoBody {
		if err := json.NewDecoder(payload).Decode(&params); err != nil {
			return apierrors.InvalidPropertyError("body", err)
		}
	}

	result, err := c.vmService.CallViewByChainID(chainID, contractHName, functionHName, params)
	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	return e.JSON(http.StatusOK, result)
}

func (c *Controller) callViewByContractName(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	contractName, err := isc.HnameFromString(e.Param("contractName"))
	if err != nil {
		return apierrors.InvalidPropertyError("contractName", err)
	}

	functionName, err := isc.HnameFromString(e.Param("functionName"))
	if err != nil {
		return apierrors.InvalidPropertyError("functionName", err)
	}

	return c.handleCallView(e, chainID, contractName, functionName, e.Request().Body)
}

func (c *Controller) callViewByHName(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	contractHName := isc.Hn(e.Param("contractHName"))
	functionHName := isc.Hn(e.Param("functionHName"))

	return c.handleCallView(e, chainID, contractHName, functionHName, e.Request().Body)
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

	contractList := make([]models.ContractInfoResponse, 0, len(contracts))

	for hName, contract := range contracts {
		contractInfo := models.ContractInfoResponse{
			Description: contract.Description,
			HName:       hName,
			Name:        contract.Name,
			ProgramHash: contract.ProgramHash,
		}

		contractList = append(contractList, contractInfo)
	}

	return e.JSON(http.StatusOK, contractList)
}
