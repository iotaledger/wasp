// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package webapi

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chain/chainutil"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/admapi"
	"github.com/iotaledger/wasp/packages/webapi/evm"
	"github.com/iotaledger/wasp/packages/webapi/info"
	"github.com/iotaledger/wasp/packages/webapi/reqstatus"
	"github.com/iotaledger/wasp/packages/webapi/request"
	"github.com/iotaledger/wasp/packages/webapi/state"
)

var log *loggerpkg.Logger

func Init(
	logger *loggerpkg.Logger,
	server echoswagger.ApiRoot,
	network peering.NetworkProvider,
	tnm peering.TrustedNetworkManager,
	registryProvider registry.Provider,
	chainsProvider chains.Provider,
	nodeProvider dkg.NodeProvider,
	shutdown admapi.ShutdownFunc,
	metrics *metricspkg.Metrics,
	w *wal.WAL,
	authConfig authentication.AuthConfiguration,
	nodeOwnerAddresses []string,
	apiCacheTTL time.Duration,
	publisherPort int,
) {
	log = logger

	server.SetRequestContentType(echo.MIMEApplicationJSON)
	server.SetResponseContentType(echo.MIMEApplicationJSON)

	pub := server.Group("public", "").SetDescription("Public endpoints")
	addWebSocketEndpoint(pub, log)

	info.AddEndpoints(pub, network, publisherPort)
	reqstatus.AddEndpoints(pub, chainsProvider.ChainProvider())
	state.AddEndpoints(pub, chainsProvider)
	evm.AddEndpoints(pub, chainsProvider, network.Self().PubKey)
	request.AddEndpoints(
		pub,
		chainsProvider.ChainProvider(),
		chainutil.GetAccountBalance,
		chainutil.HasRequestBeenProcessed,
		chainutil.CheckNonce,
		network.Self().PubKey(),
		apiCacheTTL,
		log,
	)

	adm := server.Group("admin", "").SetDescription("Admin endpoints")

	admapi.AddEndpoints(
		logger.Named("webapi/adm"),
		adm,
		network,
		tnm,
		registryProvider,
		chainsProvider,
		nodeProvider,
		shutdown,
		metrics,
		w,
		authConfig,
		nodeOwnerAddresses,
	)
	log.Infof("added web api endpoints")
}
