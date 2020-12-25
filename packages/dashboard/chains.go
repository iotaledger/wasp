package dashboard

import (
	"github.com/labstack/echo"
)

func chainsInit(e *echo.Echo, r renderer) Tab {
	tab := initChainList(e, r)
	initChain(e, r)
	initChainAccount(e, r)
	initChainBlob(e, r)
	initChainContract(e, r)
	return tab
}
