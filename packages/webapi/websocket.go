package webapi

import (
	_ "embed"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type webSocketAPI struct {
	pws *publisherws.PublisherWebSocket
}

func addWebSocketEndpoint(e echoswagger.ApiGroup, log *logger.Logger) *webSocketAPI {
	api := &webSocketAPI{
		pws: publisherws.New(log, []string{"state", "vmmsg"}),
	}

	e.GET("/chain/:chainid/ws", api.handleWebSocket)

	return api
}

func (w *webSocketAPI) handleWebSocket(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}
	return w.pws.ServeHTTP(chainID, c.Response(), c.Request())
}
