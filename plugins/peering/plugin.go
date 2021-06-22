// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	peering_udp "github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/plugins/registry"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

const (
	pluginName = "Peering"
)

var (
	log                    *logger.Logger
	defaultNetworkProvider peering_pkg.NetworkProvider // A singleton instance.
	peerNetworkConfig      coretypes.PeerNetworkConfigProvider
)

// Init is an entry point for this plugin.
func Init(suite *pairing.SuiteBn256) *node.Plugin {
	configure := func(_ *node.Plugin) {
		log = logger.NewLogger(pluginName)
		var err error
		var nodeKeyPair *key.Pair
		if nodeKeyPair, err = registry.DefaultRegistry().GetNodeIdentity(); err != nil {
			panic(err)
		}
		peerNetworkConfig, err = peering_pkg.NewStaticPeerNetworkConfigProvider(
			parameters.GetString(parameters.PeeringMyNetID),
			parameters.GetInt(parameters.PeeringPort),
			parameters.GetStringSlice(parameters.PeeringNeighbors)...,
		)
		if err != nil {
			log.Panicf("Init.peering: %w", err)
		}
		log.Infof("default peering configuration: %s", peerNetworkConfig.String())
		defaultNetworkProvider, err = peering_udp.NewNetworkProvider(
			peerNetworkConfig,
			nodeKeyPair,
			suite,
			log,
		)
		if err != nil {
			log.Panicf("Init.peering: %w", err)
		}
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

func DefaultPeerNetworkConfig() coretypes.PeerNetworkConfigProvider {
	return peerNetworkConfig
}

func GossipToNeighbors(upToNumPeers uint16, msg *peering_pkg.PeerMessage) error {
	neighbors := peerNetworkConfig.Neighbors()
	peerDomainProvider, err := defaultNetworkProvider.PeerDomain(neighbors)
	if err != nil {
		return err
	}
	peerDomainProvider.SendMsgToRandomPeers(upToNumPeers, msg)
	return nil
}
