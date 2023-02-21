package webapi

import (
	_ "embed"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/webapi/websocket"
	_ "github.com/iotaledger/wasp/packages/webapi/websocket"
)

func addWebSocketEndpoint(e echoswagger.ApiRoot, hub *websockethub.Hub, log *logger.Logger, pub *publisher.Publisher) {
	pws := websocket.NewPublisher(log, hub, []string{
		publisher.ISCEventKindNewBlock,
		publisher.ISCEventKindReceipt,
		publisher.ISCEventIssuerVM,
		publisher.ISCEventKindSmartContract,
	}, pub)

	e.GET("/ws", pws.ServeHTTP).
		SetSummary("The websocket connection service")
}
