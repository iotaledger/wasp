package chain

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/controllers/corecontracts"
)

func (c *Controller) getReceipt(e echo.Context) error {
	controllerutils.SetOperation(e, "get_receipt")
	return corecontracts.GetRequestReceipt(e, c.chainService)
}
