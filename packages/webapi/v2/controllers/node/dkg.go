package node

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
	"github.com/labstack/echo/v4"
)

func (c *Controller) generateDKG(e echo.Context) error {
	generateDKGRequest := models.DKSharesPostRequest{}

	if err := e.Bind(&generateDKGRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	publicKeys := make([]cryptolib.PublicKey, 0)

	for _, key := range generateDKGRequest.PeerPubKeys {
		publicKeys = append(publicKeys)
	}

	c.dkgService.GenerateDistributedKey()
}
