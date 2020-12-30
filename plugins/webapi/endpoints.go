package webapi

import (
	"net"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/info"
	"github.com/iotaledger/wasp/plugins/webapi/request"
	"github.com/iotaledger/wasp/plugins/webapi/state"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addEndpoints(server echoswagger.ApiRoot, adminWhitelist []net.IP) {
	server.SetRequestContentType("application/json")
	server.SetResponseContentType("application/json")

	pub := server.Group("public", "").SetDescription("Public endpoints")
	info.AddEndpoints(pub)
	request.AddEndpoints(pub)
	state.AddEndpoints(pub)

	adm := server.Group("admin", "/"+client.AdminRoutePrefix).SetDescription("Admin endpoints")
	admapi.AddEndpoints(adm, adminWhitelist)
	log.Infof("added web api endpoints")
}
