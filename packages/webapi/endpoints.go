package webapi

import (
	"net"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/webapi/admapi"
	"github.com/iotaledger/wasp/packages/webapi/blob"
	"github.com/iotaledger/wasp/packages/webapi/info"
	"github.com/iotaledger/wasp/packages/webapi/request"
	"github.com/iotaledger/wasp/packages/webapi/state"
	"github.com/pangpanglabs/echoswagger/v2"
)

var log *logger.Logger

func Init(server echoswagger.ApiRoot, adminWhitelist []net.IP) {
	log = logger.NewLogger("WebAPI")

	server.SetRequestContentType("application/json")
	server.SetResponseContentType("application/json")

	pub := server.Group("public", "").SetDescription("Public endpoints")
	blob.AddEndpoints(pub)
	info.AddEndpoints(pub)
	request.AddEndpoints(pub)
	state.AddEndpoints(pub)

	adm := server.Group("admin", "").SetDescription("Admin endpoints")
	admapi.AddEndpoints(adm, adminWhitelist)
	log.Infof("added web api endpoints")
}
