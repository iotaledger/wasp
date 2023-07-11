package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) getContracts(e echo.Context) error {
	controllerutils.SetOperation(e, "get_contracts")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	contracts, err := c.chainService.GetContracts(chainID, e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return err
	}

	contractList := make([]models.ContractInfoResponse, 0, len(contracts))

	for hName, contract := range contracts {
		contractInfo := models.ContractInfoResponse{
			HName:       hName.String(),
			Name:        contract.Name,
			ProgramHash: contract.ProgramHash.String(),
		}

		contractList = append(contractList, contractInfo)
	}

	return e.JSON(http.StatusOK, contractList)
}
