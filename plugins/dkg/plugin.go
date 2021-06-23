// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	hive_node "github.com/iotaledger/hive.go/node"
	dkg_pkg "github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/registry"
	"go.uber.org/zap"
)

const pluginName = "DKG"

var (
	defaultNode *dkg_pkg.Node // A singleton.
)

// Init is an entry point for the plugin.
func Init() *hive_node.Plugin {
	configure := func(_ *hive_node.Plugin) {
		log := logger.NewLogger(pluginName)
		registry := registry.DefaultRegistry()
		peeringProvider := peering.DefaultNetworkProvider()
		var err error
		var nodeIdentity *ed25519.KeyPair
		if nodeIdentity, err = registry.GetNodeIdentity(); err != nil {
			panic("cannot get the node key")
		}
		defaultNode = dkg_pkg.NewNode(
			nodeIdentity,
			peeringProvider,
			registry,
			log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelWarn)).Sugar(),
		)
	}
	run := func(_ *hive_node.Plugin) {
		// Nothing to run here.
	}
	return hive_node.NewPlugin(pluginName, hive_node.Enabled, configure, run)
}

// DefaultNode returns the default instance of the DKG Node Provider.
// It should be used to access all the DKG Node functions (not the DKG Initiator's).
func DefaultNode() *dkg_pkg.Node {
	return defaultNode
}
