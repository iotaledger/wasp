// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/lpp"
	"github.com/iotaledger/wasp/packages/registry"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:     "Peering",
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Params:   params,
			Provide:  provide,
			Run:      run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
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
			CoreComponent.Logger(),
		)
		if err != nil {
			CoreComponent.LogPanicf("Init.peering: %v", err)
		}
		CoreComponent.LogInfof("------------- PeeringURL = %s, PubKey = %s ------------------", ParamsPeering.PeeringURL, nodeIdentity.GetPublicKey().String())

		return networkResult{
			NetworkProvider:       netImpl,
			TrustedNetworkManager: tnmImpl,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func run() error {
	err := CoreComponent.Daemon().BackgroundWorker(
		"WaspPeering",
		deps.NetworkProvider.Run,
		daemon.PriorityPeering,
	)
	if err != nil {
		panic(err)
	}

	return nil
}
