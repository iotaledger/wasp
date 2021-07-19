package webapi

import (
	"net"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/admapi"
	"github.com/iotaledger/wasp/packages/webapi/blob"
	"github.com/iotaledger/wasp/packages/webapi/info"
	"github.com/iotaledger/wasp/packages/webapi/reqstatus"
	"github.com/iotaledger/wasp/packages/webapi/request"
	"github.com/iotaledger/wasp/packages/webapi/state"
	"github.com/iotaledger/wasp/packages/webapi/webapiutil"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

var log *logger.Logger

func Init(server echoswagger.ApiRoot, adminWhitelist []net.IP, network peering.NetworkProvider, tnm peering.TrustedNetworkManager, registryProvider registry.Provider, chainsProvider chains.Provider) {
	log = logger.NewLogger("WebAPI")

	server.SetRequestContentType(echo.MIMEApplicationJSON)
	server.SetResponseContentType(echo.MIMEApplicationJSON)

	pub := server.Group("public", "").SetDescription("Public endpoints")
	blob.AddEndpoints(pub, func() registry.BlobCache { return registryProvider() })
	info.AddEndpoints(pub)
	reqstatus.AddEndpoints(pub)
	request.AddEndpoints(pub, webapiutil.GetChain, webapiutil.GetAccountBalance)
	state.AddEndpoints(pub)

	adm := server.Group("admin", "").SetDescription("Admin endpoints")
	admapi.AddEndpoints(adm, adminWhitelist, network, tnm, registryProvider, chainsProvider)
	log.Infof("added web api endpoints")
}
