// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
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
	authConfig authentication.AuthConfiguration,
	nodeOwnerAddresses []string,
) {
	initLogger()

	claimValidator := func(claims *authentication.WaspClaims) bool {
		// The API will be accessible if the token has an 'API' claim
		return claims.HasPermission(permissions.API)
	}

	authentication.AddAuthentication(adm.EchoGroup(), registryProvider, authConfig, claimValidator)
	addShutdownEndpoint(adm, shutdown)
	addNodeOwnerEndpoints(adm, registryProvider, nodeOwnerAddresses)
	addChainRecordEndpoints(adm, registryProvider)
	addChainMetricsEndpoints(adm, chainsProvider)
	addChainEndpoints(adm, &chainWebAPI{
		registry:   registryProvider,
		chains:     chainsProvider,
		network:    network,
		allMetrics: metrics,
		w:          w,
	})
	addDKSharesEndpoints(adm, registryProvider, nodeProvider)
	addPeeringEndpoints(adm, network, tnm)
}
