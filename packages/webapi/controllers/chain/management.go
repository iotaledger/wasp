package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) activateChain(e echo.Context) error {
	controllerutils.SetOperation(e, "activate_chain")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	if err := c.chainService.ActivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) deactivateChain(e echo.Context) error {
	controllerutils.SetOperation(e, "deactivate_chain")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	if err := c.chainService.DeactivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) setChainRecord(e echo.Context) error {
	controllerutils.SetOperation(e, "set_chain_record")
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	// No need to validate the chain existence here (like above), as the service will create a chain record if it does not exist.

	var request models.ChainRecord
	if err := e.Bind(&request); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	record := registry.NewChainRecord(chainID, request.IsActive, []*cryptolib.PublicKey{})

	for _, publicKeyStr := range request.AccessNodes {
		publicKey, err := cryptolib.PublicKeyFromString(publicKeyStr)
		if err != nil {
			return apierrors.InvalidPropertyError("accessNode", err)
		}

		record.AccessNodes = append(record.AccessNodes, publicKey)
	}

	if err := c.chainService.SetChainRecord(record); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}
