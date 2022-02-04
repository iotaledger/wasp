// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package webapi

import (
	"net"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/admapi"
	"github.com/iotaledger/wasp/packages/webapi/info"
	"github.com/iotaledger/wasp/packages/webapi/reqstatus"
	"github.com/iotaledger/wasp/packages/webapi/request"
	"github.com/iotaledger/wasp/packages/webapi/state"
	"github.com/iotaledger/wasp/packages/webapi/webapiutil"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

var log *logger.Logger

func Init(
	server echoswagger.ApiRoot,
	adminWhitelist []net.IP,
	network peering.NetworkProvider,
	tnm peering.TrustedNetworkManager,
	registryProvider registry.Provider,
	chainsProvider chains.Provider,
	nodeProvider dkg.NodeProvider,
	shutdown admapi.ShutdownFunc,
	metrics *metricspkg.Metrics,
	w *wal.WAL,
) {
	log = logger.NewLogger("WebAPI")

	server.SetRequestContentType(echo.MIMEApplicationJSON)
	server.SetResponseContentType(echo.MIMEApplicationJSON)

	pub := server.Group("public", "").SetDescription("Public endpoints")
	addWebSocketEndpoint(pub, log)

	info.AddEndpoints(pub, network)
	reqstatus.AddEndpoints(pub, chainsProvider.ChainProvider())
	state.AddEndpoints(pub, chainsProvider)
	request.AddEndpoints(
		pub,
		chainsProvider.ChainProvider(),
		webapiutil.GetAccountBalance,
		webapiutil.HasRequestBeenProcessed,
		network.Self().PubKey(),
		time.Duration(parameters.GetInt(parameters.OffledgerAPICacheTTL))*time.Second,
		log,
	)

	adm := server.Group("admin", "").SetDescription("Admin endpoints")
	admapi.AddEndpoints(
		adm,
		adminWhitelist,
		network,
		tnm,
		registryProvider,
		chainsProvider,
		nodeProvider,
		shutdown,
		metrics,
		w,
	)
	log.Infof("added web api endpoints")
}
