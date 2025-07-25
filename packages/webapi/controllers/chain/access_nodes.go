package chain

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
)

func (c *Controller) addAccessNode(e echo.Context) error {
	controllerutils.SetOperation(e, "add_access_node")
	chain, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	peer := e.Param(params.ParamPeer)
	if peer == "" {
		return errors.New("no peer provided")
	}

	if err := c.nodeService.AddAccessNode(chain.ID(), peer); err != nil {
		if errors.Is(err, interfaces.ErrPeerNotFound) {
			return apierrors.PeerNameNotFoundError(peer)
		}

		return err
	}

	return e.NoContent(http.StatusCreated)
}

func (c *Controller) removeAccessNode(e echo.Context) error {
	controllerutils.SetOperation(e, "remove_access_node")
	chain, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	peer := e.Param(params.ParamPeer)
	if peer == "" {
		return errors.New("no peer provided")
	}

	if err := c.nodeService.DeleteAccessNode(chain.ID(), peer); err != nil {
		return err
	}

	return e.NoContent(http.StatusOK)
}
