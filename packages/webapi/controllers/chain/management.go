package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
)

func (c *Controller) activateChain(e echo.Context) error {
	controllerutils.SetOperation(e, "activate_chain")
	chainID, err := controllerutils.ChainIDFromParams(e)
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
	chain, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	if err := c.chainService.DeactivateChain(chain.ID()); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) rotateChain(e echo.Context) error {
	controllerutils.SetOperation(e, "rotate_chain")

	var request models.RotateChainRequest
	if err := e.Bind(&request); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	var rotateToAddress *iotago.Address
	if request.RotateToAddress == nil || *request.RotateToAddress == "" {
		rotateToAddress = nil
	} else {
		rotateToAddress = iotago.MustAddressFromHex(*request.RotateToAddress)
	}

	if err := c.chainService.RotateTo(e.Request().Context(), rotateToAddress); err != nil {
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
