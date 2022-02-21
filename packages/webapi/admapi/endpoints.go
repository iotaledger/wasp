// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/pangpanglabs/echoswagger/v2"
)

var log *logger.Logger
var jwtAuth *authentication.JWTAuth

func initLogger() {
	log = logger.NewLogger("webapi/adm")
}

func AddEndpoints(
	adm echoswagger.ApiGroup,
	network peering.NetworkProvider,
	tnm peering.TrustedNetworkManager,
	registryProvider registry.Provider,
	chainsProvider chains.Provider,
	nodeProvider dkg.NodeProvider,
	shutdown ShutdownFunc,
	metrics *metricspkg.Metrics,
	w *wal.WAL,
) {
	initLogger()

	authentication.AddAuthentication(adm.EchoGroup(), registryProvider, parameters.WebAPIAuth, "api")
	addShutdownEndpoint(adm, shutdown)
	addNodeOwnerEndpoints(adm, registryProvider)
	addChainRecordEndpoints(adm, registryProvider)
	addChainMetricsEndpoints(adm, chainsProvider)
	addChainEndpoints(adm, registryProvider, chainsProvider, network, metrics, w)
	addDKSharesEndpoints(adm, registryProvider, nodeProvider)
	addPeeringEndpoints(adm, network, tnm)
}
