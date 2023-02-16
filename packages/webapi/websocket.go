package webapi

import (
	_ "embed"

	"github.com/iotaledger/hive.go/core/websockethub"
	"github.com/iotaledger/wasp/packages/publisher"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
)

func addWebSocketEndpoint(e echoswagger.ApiRoot, hub *websockethub.Hub, log *logger.Logger, pub *publisher.Publisher) {
	pws := publisherws.New(log, hub, []string{
		publisher.ISCEventKindNewBlock,
		publisher.ISCEventKindReceipt,
		publisher.ISCEventIssuerVM,
	}, pub)

	e.GET("/ws", pws.ServeHTTP)
}
