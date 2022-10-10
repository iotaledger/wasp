package chain

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"

	"github.com/iotaledger/wasp/packages/isc"
)

func (c *Controller) activateChain(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	if err := c.chainService.ActivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) deactivateChain(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	if err := c.chainService.DeactivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) saveChain(e echo.Context) error {
	var saveChainRequest models.SaveChainRecordRequest
	if err := e.Bind(&saveChainRequest); err != nil {
		return err
	}

	chainID, err := isc.ChainIDFromString(saveChainRequest.ChainID)
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	if err := c.chainService.SaveChainRecord(chainID, saveChainRequest.Active); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}
