package chain

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/params"
)

func (c *Controller) handleWebSocket(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	return c.webSocketHandler.ServeHTTP(chainID, e.Response(), e.Request())
}
