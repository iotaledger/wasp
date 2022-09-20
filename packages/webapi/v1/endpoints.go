// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/iotaledger/wasp/packages/webapi/v1/admapi"
	"github.com/iotaledger/wasp/packages/webapi/v1/evm"
	"github.com/iotaledger/wasp/packages/webapi/v1/info"
	"github.com/iotaledger/wasp/packages/webapi/v1/reqstatus"
	"github.com/iotaledger/wasp/packages/webapi/v1/request"
	"github.com/iotaledger/wasp/packages/webapi/v1/state"
)

var log *loggerpkg.Logger

func Init(
	logger *loggerpkg.Logger,
	server echoswagger.ApiRoot,
	network peering.NetworkProvider,
	tnm peering.TrustedNetworkManager,
	userManager *users.UserManager,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	chainsProvider chains.Provider,
	nodeProvider dkg.NodeProvider,
	shutdown admapi.ShutdownFunc,
	nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics,
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
		userManager,
		chainRecordRegistryProvider,
		dkShareRegistryProvider,
		nodeIdentityProvider,
		chainsProvider,
		nodeProvider,
		shutdown,
		nodeConnectionMetrics,
		authConfig,
		nodeOwnerAddresses,
	)
	log.Infof("added web api endpoints")
}
