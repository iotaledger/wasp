// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/jwt_auth"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util/auth"
	"github.com/pangpanglabs/echoswagger/v2"
)

var log *logger.Logger
var jwtAuth *jwt_auth.JWTAuth

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
) {
	initLogger()

	var config auth.BaseAuthConfiguration
	parameters.GetStruct(parameters.WebAPIAuth, &config)
	auth.AddAuthenticationWebAPI(adm, config)

	addShutdownEndpoint(adm, shutdown)
	addNodeOwnerEndpoints(adm, registryProvider)
	addChainRecordEndpoints(adm, registryProvider)
	addChainMetricsEndpoints(adm, chainsProvider)
	addChainEndpoints(adm, registryProvider, chainsProvider, network, metrics)
	addDKSharesEndpoints(adm, registryProvider, nodeProvider)
	addPeeringEndpoints(adm, network, tnm)
}

// allow only if the remote address is private or in whitelist
// TODO this is a very basic/limited form of protection
