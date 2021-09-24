package dashboard

import (
	_ "embed"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
	"github.com/labstack/echo/v4"
)

//go:embed templates/websocket.tmpl
var tplWebSocket string

func (d *Dashboard) webSocketInit(e *echo.Echo) {
	pws := publisherws.New(d.log, []string{"state"})

	route := e.GET("/chain/:chainid/ws", func(c echo.Context) error {
		chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
		if err != nil {
			return err
		}
		return pws.ServeHTTP(chainID, c.Response(), c.Request())
	})
	route.Name = "chainWebSocket"
}
