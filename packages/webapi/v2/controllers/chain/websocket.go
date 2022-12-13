package chain

import (
	"github.com/iotaledger/wasp/packages/webapi/v2/params"
	"github.com/labstack/echo/v4"
)

func (c *Controller) handleWebSocket(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	return c.webSocketHandler.ServeHTTP(chainID, e.Response(), e.Request())
}
