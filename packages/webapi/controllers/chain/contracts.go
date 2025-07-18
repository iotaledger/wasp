package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
)

func (c *Controller) getContracts(e echo.Context) error {
	controllerutils.SetOperation(e, "get_contracts")

	contracts, err := c.chainService.GetContracts(e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return err
	}

	contractList := make([]models.ContractInfoResponse, 0, len(contracts))

	for _, contract := range contracts {
		hName, contract := contract.Unpack()

		contractInfo := models.ContractInfoResponse{
			HName: hName.String(),
			Name:  contract.Name,
		}

		contractList = append(contractList, contractInfo)
	}

	return e.JSON(http.StatusOK, contractList)
}
