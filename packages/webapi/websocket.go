package webapi

import (
	_ "embed"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/webapi/websocket"
)

func addWebSocketEndpoint(e echoswagger.ApiRoot, websocketPublisher *websocket.Service) {
	e.GET("/ws", websocketPublisher.ServeHTTP).
		SetSummary("The websocket connection service")
}
