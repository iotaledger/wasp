// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/dkg"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
)

var log *loggerpkg.Logger

func AddEndpoints(
	logger *loggerpkg.Logger,
	adm echoswagger.ApiGroup,
	network peering.NetworkProvider,
	tnm peering.TrustedNetworkManager,
	userManager *users.UserManager,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	consensusJournalRegistryProvider journal.Provider,
	chainsProvider chains.Provider,
	nodeProvider dkg.NodeProvider,
	shutdown ShutdownFunc,
	metrics *metricspkg.Metrics,
	authConfig authentication.AuthConfiguration,
	nodeOwnerAddresses []string,
) {
	log = logger

	claimValidator := func(claims *authentication.WaspClaims) bool {
		// The API will be accessible if the token has an 'API' claim
		return claims.HasPermission(permissions.API)
	}

	authentication.AddAuthentication(adm.EchoGroup(), userManager, nodeIdentityProvider, authConfig, claimValidator)
	addShutdownEndpoint(adm, shutdown)
	addNodeOwnerEndpoints(adm, nodeIdentityProvider, nodeOwnerAddresses)
	addChainRecordEndpoints(adm, chainRecordRegistryProvider)
	addChainMetricsEndpoints(adm, chainsProvider)
	addChainEndpoints(adm, &chainWebAPI{
		chainRecordRegistryProvider:      chainRecordRegistryProvider,
		dkShareRegistryProvider:          dkShareRegistryProvider,
		nodeIdentityProvider:             nodeIdentityProvider,
		consensusJournalRegistryProvider: consensusJournalRegistryProvider,
		chains:                           chainsProvider,
		network:                          network,
		allMetrics:                       metrics,
		w:                                w,
	})
	addDKSharesEndpoints(adm, dkShareRegistryProvider, nodeProvider)
	addPeeringEndpoints(adm, chainRecordRegistryProvider, network, tnm)
	addAccessNodesEndpoints(adm, chainRecordRegistryProvider, tnm)
}
