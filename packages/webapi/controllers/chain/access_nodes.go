package chain

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

func decodeAccessNodeRequest(e echo.Context) (isc.ChainID, string, error) {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return isc.EmptyChainID(), "", err
	}

	peer := e.Param("peer")
	if peer == "" {
		return isc.EmptyChainID(), "", errors.New("no peer provided")
	}

	return chainID, peer, nil
}

func (c *Controller) addAccessNode(e echo.Context) error {
	chainID, peer, err := decodeAccessNodeRequest(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	if err := c.nodeService.AddAccessNode(chainID, peer); err != nil {
		if errors.Is(err, interfaces.ErrPeerNotFound) {
			return apierrors.PeerNameNotFoundError(peer)
		}

		return err
	}

	return e.NoContent(http.StatusCreated)
}

func (c *Controller) removeAccessNode(e echo.Context) error {
	chainID, peer, err := decodeAccessNodeRequest(e)
	if err != nil {
		return err
	}

	if !c.chainService.HasChain(chainID) {
		return apierrors.ChainNotFoundError(chainID.String())
	}

	if err := c.nodeService.DeleteAccessNode(chainID, peer); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}
