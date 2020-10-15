package webapi

import (
	"net"

	"github.com/iotaledger/wasp/plugins/webapi/admapi"
	"github.com/iotaledger/wasp/plugins/webapi/info"
	"github.com/iotaledger/wasp/plugins/webapi/request"
	"github.com/iotaledger/wasp/plugins/webapi/state"
)

func addEndpoints(adminWhitelist []net.IP) {
	info.AddEndpoints(Server)
	request.AddEndpoints(Server)
	state.AddEndpoints(Server)
	admapi.AddEndpoints(Server, adminWhitelist)
	log.Infof("added web api endpoints")
}
