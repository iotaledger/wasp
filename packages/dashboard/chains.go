package dashboard

import (
	"github.com/labstack/echo/v4"
)

func (d *Dashboard) chainsInit(e *echo.Echo, r renderer) Tab {
	tab := d.initChainList(e, r)
	d.initChain(e, r)
	d.initChainAccount(e, r)
	d.initChainBlob(e, r)
	d.initChainContract(e, r)
	d.initChainBlock(e, r)
	return tab
}
