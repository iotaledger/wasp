// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	peering_lpp "github.com/iotaledger/wasp/packages/peering/lpp"
	"github.com/iotaledger/wasp/plugins/registry"
)

const (
	pluginName = "Peering"
)

var (
	log                          *logger.Logger
	defaultNetworkProvider       peering_pkg.NetworkProvider       // A singleton instance.
	defaultTrustedNetworkManager peering_pkg.TrustedNetworkManager // A singleton instance.
)

// Init is an entry point for this plugin.
func Init() *node.Plugin {
	configure := func(_ *node.Plugin) {
		log = logger.NewLogger(pluginName)
		var err error
		var nodeKeyPair *cryptolib.KeyPair
		if nodeKeyPair, err = registry.DefaultRegistry().GetNodeIdentity(); err != nil {
			panic(err)
		}
		if err != nil {
			log.Panicf("Init.peering: %v", err)
		}
		netID := parameters.GetString(parameters.PeeringMyNetID)
		netImpl, tnmImpl, err := peering_lpp.NewNetworkProvider(
			netID,
			parameters.GetInt(parameters.PeeringPort),
			nodeKeyPair,
			registry.DefaultRegistry(),
			log,
		)
		if err != nil {
			log.Panicf("Init.peering: %v", err)
		}
		defaultNetworkProvider = netImpl
		defaultTrustedNetworkManager = tnmImpl
		log.Infof("------------- NetID is %s ------------------", netID)
	}
	run := func(_ *node.Plugin) {
		err := daemon.BackgroundWorker(
			"WaspPeering",
			defaultNetworkProvider.Run,
			parameters.PriorityPeering,
		)
		if err != nil {
			panic(err)
		}
	}
	return node.NewPlugin(PluginName, nil, node.Enabled, configure, run)
}

// DefaultNetworkProvider returns the default network provider implementation.
func DefaultNetworkProvider() peering_pkg.NetworkProvider {
	return defaultNetworkProvider
}

func DefaultTrustedNetworkManager() peering_pkg.TrustedNetworkManager {
	return defaultTrustedNetworkManager
}
