package webapi

import (
	"net"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/admapi"
	"github.com/iotaledger/wasp/packages/webapi/blob"
	"github.com/iotaledger/wasp/packages/webapi/info"
	"github.com/iotaledger/wasp/packages/webapi/reqstatus"
	"github.com/iotaledger/wasp/packages/webapi/request"
	"github.com/iotaledger/wasp/packages/webapi/state"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

var log *logger.Logger

func getChain(chainID *coretypes.ChainID) chain.ChainCore {
	return chain.ChainCore(chains.AllChains().Get(chainID))
}

func Init(server echoswagger.ApiRoot, adminWhitelist []net.IP) {
	log = logger.NewLogger("WebAPI")

	server.SetRequestContentType(echo.MIMEApplicationJSON)
	server.SetResponseContentType(echo.MIMEApplicationJSON)

	pub := server.Group("public", "").SetDescription("Public endpoints")
	blob.AddEndpoints(pub)
	info.AddEndpoints(pub)
	reqstatus.AddEndpoints(pub)
	// getChain is a workaround for testing (to use MockedChainCore), maybe there is a cleaner way
	request.AddEndpoints(pub, getChain, log)
	// request.AddEndpoints(pub, chains.AllChains().Get)
	state.AddEndpoints(pub)

	adm := server.Group("admin", "").SetDescription("Admin endpoints")
	admapi.AddEndpoints(adm, adminWhitelist)
	log.Infof("added web api endpoints")
}
