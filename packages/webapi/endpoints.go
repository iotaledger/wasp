// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package webapi

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/prometheus/tsdb/wal"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/dkg"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
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
	userManager *users.UserManager,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	chainsProvider chains.Provider,
	consensusJournalRegistryProvider journal.Provider,
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
		userManager,
		chainRecordRegistryProvider,
		dkShareRegistryProvider,
		nodeIdentityProvider,
		consensusJournalRegistryProvider,
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
