package dashboard

import (
	_ "embed"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
)

//go:embed templates/websocket.tmpl
var tplWebSocket string

func (d *Dashboard) webSocketInit(e *echo.Echo) {
	pws := publisherws.New(d.log, []string{"state"})

	route := e.GET("/chain/:chainid/ws", func(c echo.Context) error {
		chainID, err := isc.ChainIDFromString(c.Param("chainid"))
		if err != nil {
			return err
		}

		return pws.ServeHTTP(chainID, c.Response(), c.Request())
	})
	route.Name = "chainWebSocket"
}
