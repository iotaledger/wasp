package chain

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
)

func (c *Controller) handleJSONRPC(e echo.Context) error {
	controllerutils.SetOperation(e, "evm_json_rpc")
	return c.evmService.HandleJSONRPC(e.Request(), e.Response())
}

func (c *Controller) handleWebsocket(e echo.Context) error {
	controllerutils.SetOperation(e, "evm_websocket")

	ctx := e.Echo().Server.BaseContext(nil)
	return c.evmService.HandleWebsocket(ctx, e)
}
