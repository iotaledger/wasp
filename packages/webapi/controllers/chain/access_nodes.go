package chain

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func decodeAccessNodeRequest(e echo.Context) (isc.ChainID, *cryptolib.PublicKey, error) {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return isc.EmptyChainID(), nil, err
	}

	publicKey, err := params.DecodePublicKey(e)
	if err != nil {
		return isc.EmptyChainID(), nil, err
	}

	return chainID, publicKey, nil
}

func (c *Controller) addAccessNode(e echo.Context) error {
	chainID, publicKey, err := decodeAccessNodeRequest(e)
	if err != nil {
		return err
	}

	if err := c.nodeService.AddAccessNode(chainID, publicKey); err != nil {
		if errors.Is(err, interfaces.ErrPeerNotFound) {
			return apierrors.PeerNotFoundError(publicKey)
		}

		return err
	}

	return e.NoContent(http.StatusCreated)
}

func (c *Controller) removeAccessNode(e echo.Context) error {
	chainID, publicKey, err := decodeAccessNodeRequest(e)
	if err != nil {
		return err
	}

	if err := c.nodeService.DeleteAccessNode(chainID, publicKey); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}
