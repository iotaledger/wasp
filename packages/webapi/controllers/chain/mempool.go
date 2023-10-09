package chain

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
)

func (c *Controller) getMempoolContents(e echo.Context) error {
	controllerutils.SetOperation(e, "get_mempool_contents")
	ch, _, err := controllerutils.ChainFromParams(e, c.chainService)
	if err != nil {
		return err
	}
	return e.Stream(http.StatusOK, "application/octet-stream", ch.GetMempoolContents())
}
