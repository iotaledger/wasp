// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	peering_udp "github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/plugins/registry"
)

const (
	pluginName = "Peering"
)

var (
	log                          *logger.Logger
	defaultNetworkProvider       peering_pkg.NetworkProvider         // A singleton instance.
	defaultTrustedNetworkManager peering_pkg.TrustedNetworkManager   // A singleton instance.
	peerNetworkConfig            coretypes.PeerNetworkConfigProvider // TODO: Remove.
)

// Init is an entry point for this plugin.
func Init() *node.Plugin {
	configure := func(_ *node.Plugin) {
		log = logger.NewLogger(pluginName)
		var err error
		var nodeKeyPair *ed25519.KeyPair
		if nodeKeyPair, err = registry.DefaultRegistry().GetNodeIdentity(); err != nil {
			panic(err)
		}
		peerNetworkConfig, err = peering_pkg.NewStaticPeerNetworkConfigProvider( // TODO: Remove.
			parameters.GetString(parameters.PeeringMyNetId),
			parameters.GetInt(parameters.PeeringPort),
			parameters.GetStringSlice(parameters.PeeringNeighbors)..., // Unregister the parameter?
		)
		if err != nil {
			log.Panicf("Init.peering: %w", err)
		}
		log.Infof("default peering configuration: %s", peerNetworkConfig.String())
		netImpl, err := peering_udp.NewNetworkProvider(
			parameters.GetString(parameters.PeeringMyNetId),
			parameters.GetInt(parameters.PeeringPort),
			*nodeKeyPair,
			registry.DefaultRegistry(),
			log,
		)
		if err != nil {
			log.Panicf("Init.peering: %w", err)
		}
		defaultNetworkProvider = netImpl
		defaultTrustedNetworkManager = netImpl
		log.Infof("------------- NetID is %s ------------------", peerNetworkConfig.OwnNetID())
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
	return node.NewPlugin(pluginName, node.Enabled, configure, run)
}

// DefaultNetworkProvider returns the default network provider implementation.
func DefaultNetworkProvider() peering_pkg.NetworkProvider {
	return defaultNetworkProvider
}

func DefaultTrustedNetworkManager() peering_pkg.TrustedNetworkManager {
	return defaultTrustedNetworkManager
}

func DefaultPeerNetworkConfig() coretypes.PeerNetworkConfigProvider { // TODO: Remove.
	return peerNetworkConfig
}
