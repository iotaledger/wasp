package chain

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/labstack/echo/v4"
)

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
