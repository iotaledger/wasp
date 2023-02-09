package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) activateChain(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if err := c.chainService.ActivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) deactivateChain(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	if err := c.chainService.DeactivateChain(chainID); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) setChainRecord(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	var request models.ChainRecord
	if err := e.Bind(&request); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	record := registry.NewChainRecord(chainID, request.IsActive, []*cryptolib.PublicKey{})

	for _, publicKeyStr := range request.AccessNodes {
		publicKey, err := cryptolib.NewPublicKeyFromString(publicKeyStr)
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
