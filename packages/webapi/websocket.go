// Package webapi provides types and methods for implementing the wasp http and websocket api
package webapi

import (
	_ "embed"
	"fmt"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/v2/packages/webapi/websocket"
)

func addWebSocketEndpoint(e echoswagger.ApiRoot, websocketPublisher *websocket.Service) {
	e.GET(fmt.Sprintf("/v%d/ws", APIVersion), websocketPublisher.ServeHTTP).
		SetSummary("The websocket connection service")
}
