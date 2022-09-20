package v1

import (
	_ "embed"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
)

type webSocketAPI struct {
	pws *publisherws.PublisherWebSocket
}

func addWebSocketEndpoint(e echoswagger.ApiGroup, log *logger.Logger) *webSocketAPI {
	api := &webSocketAPI{
		pws: publisherws.New(log, []string{"state", "contract"}),
	}

	e.GET("/chain/:chainID/ws", api.handleWebSocket).
		AddParamPath("", "chainID", "ChainID (bech32-encoded)")

	return api
}

func (w *webSocketAPI) handleWebSocket(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return err
	}
	return w.pws.ServeHTTP(chainID, c.Response(), c.Request())
}
