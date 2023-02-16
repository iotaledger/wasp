package webapi

import (
	_ "embed"

	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
)

type webSocketAPI struct {
	pws *publisherws.PublisherWebSocket
}

func addWebSocketEndpoint(e echoswagger.ApiRoot, hub *websockethub.Hub, log *logger.Logger) *webSocketAPI {
	api := &webSocketAPI{
		pws: publisherws.New(log, hub, []string{
			publisher.ISCEventKindNewBlock,
			publisher.ISCEventKindReceipt,
			publisher.ISCEventIssuerVM,
		}),
	}

	e.Echo().GET("/ws", api.handleWebSocket)

	return api
}

func (w *webSocketAPI) handleWebSocket(c echo.Context) error {
	return w.pws.ServeHTTP(c.Response(), c.Request())
}
