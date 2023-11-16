package chain

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
)

func (c *Controller) handleJSONRPC(e echo.Context) error {
	controllerutils.SetOperation(e, "evm_json_rpc")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	return c.evmService.HandleJSONRPC(chainID, e.Request(), e.Response())
}

func (c *Controller) handleWebsocket(e echo.Context) error {
	controllerutils.SetOperation(e, "evm_websocket")
	chainID, err := controllerutils.ChainIDFromParams(e, c.chainService)
	if err != nil {
		return err
	}

	ctx := e.Echo().Server.BaseContext(nil)
	return c.evmService.HandleWebsocket(ctx, chainID, e)
}
