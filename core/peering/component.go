// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/parameters"
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

	DefaultNetworkProvider peering.NetworkProvider `name:"defaultNetworkProvider"`
}

func provide(c *dig.Container) error {
	type networkDeps struct {
		dig.In

		DefaultRegistry registry.Registry
	}

	type networkResult struct {
		dig.Out

		DefaultNetworkProvider       peering.NetworkProvider       `name:"defaultNetworkProvider"`
		DefaultTrustedNetworkManager peering.TrustedNetworkManager `name:"defaultTrustedNetworkManager"`
	}

	if err := c.Provide(func(deps networkDeps) networkResult {
		netImpl, tnmImpl, err := lpp.NewNetworkProvider(
			ParamsPeering.NetID,
			ParamsPeering.Port,
			deps.DefaultRegistry.GetNodeIdentity(),
			deps.DefaultRegistry,
			CoreComponent.Logger(),
		)
		if err != nil {
			CoreComponent.LogPanicf("Init.peering: %v", err)
		}
		CoreComponent.LogInfof("------------- NetID is %s ------------------", ParamsPeering.NetID)

		return networkResult{
			DefaultNetworkProvider:       netImpl,
			DefaultTrustedNetworkManager: tnmImpl,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func run() error {
	err := CoreComponent.Daemon().BackgroundWorker(
		"WaspPeering",
		deps.DefaultNetworkProvider.Run,
		parameters.PriorityPeering,
	)
	if err != nil {
		panic(err)
	}

	return nil
}
