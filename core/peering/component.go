// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/core/registry"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	peering_lpp "github.com/iotaledger/wasp/packages/peering/lpp"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "Peering",
			Params:    params,
			Configure: configure,
			Run:       run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent

	defaultNetworkProvider       peering_pkg.NetworkProvider
	defaultTrustedNetworkManager peering_pkg.TrustedNetworkManager
)

func configure() error {
	netID := ParamsPeering.NetID
	netImpl, tnmImpl, err := peering_lpp.NewNetworkProvider(
		netID,
		ParamsPeering.Port,
		registry.DefaultRegistry().GetNodeIdentity(),
		registry.DefaultRegistry(),
		CoreComponent.Logger(),
	)
	if err != nil {
		CoreComponent.LogPanicf("Init.peering: %v", err)
	}
	defaultNetworkProvider = netImpl
	defaultTrustedNetworkManager = tnmImpl
	CoreComponent.LogInfof("------------- NetID is %s ------------------", netID)

	return nil
}

func run() error {
	err := CoreComponent.Daemon().BackgroundWorker(
		"WaspPeering",
		defaultNetworkProvider.Run,
		parameters.PriorityPeering,
	)
	if err != nil {
		panic(err)
	}

	return nil
}

// DefaultNetworkProvider returns the default network provider implementation.
func DefaultNetworkProvider() peering_pkg.NetworkProvider {
	return defaultNetworkProvider
}

func DefaultTrustedNetworkManager() peering_pkg.TrustedNetworkManager {
	return defaultTrustedNetworkManager
}
