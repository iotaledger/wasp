// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"

	"github.com/iotaledger/wasp/v2/packages/daemon"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/peering/lpp"
	"github.com/iotaledger/wasp/v2/packages/registry"
)

func init() {
	Component = &app.Component{
		Name:     "Peering",
		DepsFunc: func(cDeps dependencies) { deps = cDeps },
		Params:   params,
		Provide:  provide,
		Run:      run,
	}
}

var (
	Component *app.Component
	deps      dependencies
)

type dependencies struct {
	dig.In

	NetworkProvider peering.NetworkProvider `name:"networkProvider"`
}

func provide(c *dig.Container) error {
	type networkDeps struct {
		dig.In

		NodeIdentityProvider         registry.NodeIdentityProvider
		TrustedPeersRegistryProvider registry.TrustedPeersRegistryProvider
		PeeringMetricsProvider       *metrics.PeeringMetricsProvider
	}

	type networkResult struct {
		dig.Out

		NetworkProvider       peering.NetworkProvider       `name:"networkProvider"`
		TrustedNetworkManager peering.TrustedNetworkManager `name:"trustedNetworkManager"`
	}

	if err := c.Provide(func(deps networkDeps) networkResult {
		nodeIdentity := deps.NodeIdentityProvider.NodeIdentity()
		netImpl, tnmImpl, err := lpp.NewNetworkProvider(
			ParamsPeering.PeeringURL,
			ParamsPeering.Port,
			nodeIdentity,
			deps.TrustedPeersRegistryProvider,
			deps.PeeringMetricsProvider,
			Component.Logger,
		)
		if err != nil {
			Component.LogPanicf("Init.peering: %v", err)
		}
		Component.LogInfof("------------- PeeringURL = %s, PubKey = %s ------------------", ParamsPeering.PeeringURL, nodeIdentity.GetPublicKey().String())

		return networkResult{
			NetworkProvider:       netImpl,
			TrustedNetworkManager: tnmImpl,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}

func run() error {
	err := Component.Daemon().BackgroundWorker(
		"WaspPeering",
		deps.NetworkProvider.Run,
		daemon.PriorityPeering,
	)
	if err != nil {
		panic(err)
	}

	return nil
}
