// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
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

	claimValidator := func(claims *authentication.WaspClaims) bool {
		// The API will be accessible if the token has an 'API' claim
		return claims.HasPermission(permissions.API)
	}

	authentication.AddAuthentication(adm.EchoGroup(), registryProvider, parameters.WebAPIAuth, claimValidator)
	addShutdownEndpoint(adm, shutdown)
	addNodeOwnerEndpoints(adm, registryProvider)
	addChainRecordEndpoints(adm, registryProvider)
	addChainMetricsEndpoints(adm, chainsProvider)
	addChainEndpoints(adm, registryProvider, chainsProvider, network, metrics, w)
	addDKSharesEndpoints(adm, registryProvider, nodeProvider)
	addPeeringEndpoints(adm, network, tnm)
}
