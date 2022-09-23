package chain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"

	"github.com/labstack/echo/v4"
)

func (c *ChainController) handleCallView(e echo.Context, chainID *isc.ChainID, contractHName, functionHName isc.Hname, payload io.ReadCloser) error {
	var params dict.Dict
	if payload != http.NoBody {
		if err := json.NewDecoder(payload).Decode(&params); err != nil {
			return httperrors.BadRequest("Invalid request body")
		}
	}

	result, err := c.vmService.CallViewByChainID(chainID, contractHName, functionHName, params)
	if err != nil {
		return httperrors.ServerError(fmt.Sprintf("View call failed: %v", err))
	}

	return e.JSON(http.StatusOK, result)
}

func (c *ChainController) callViewByContractName(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain ID: %+v", e.Param("chainID")))
	}

	contractHName, err := isc.HnameFromString(e.Param("contractName"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid contract ID: %+v", e.Param("contractName")))
	}

	functionHName, err := isc.HnameFromString(e.Param("functionName"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid hname ID: %+v", e.Param("functionName")))
	}

	return c.handleCallView(e, chainID, contractHName, functionHName, e.Request().Body)
}

func (c *ChainController) callViewByHName(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain ID: %+v", e.Param("chainID")))
	}

	contractHName := isc.Hn(e.Param("contractHName"))
	functionHName := isc.Hn(e.Param("functionHName"))

	return c.handleCallView(e, chainID, contractHName, functionHName, e.Request().Body)
}

func (c *ChainController) getContracts(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return err
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
