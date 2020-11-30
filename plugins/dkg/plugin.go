package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/iotaledger/hive.go/logger"
	hive_node "github.com/iotaledger/hive.go/node"
	dkg_pkg "github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/peering"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

const pluginName = "DKG"

var (
	defaultNode *dkg_pkg.Node // A singleton.
)

// Init is an entry point for the plugin.
func Init(suite rabin_dkg.Suite) *hive_node.Plugin {
	configure := func(_ *hive_node.Plugin) {
		logger := logger.NewLogger(pluginName)
		registry := registry.DefaultRegistry()
		peeringProvider := peering.DefaultNetworkProvider()
		defaultNode = dkg_pkg.Init(
			nil, // TODO: SecKey
			nil, // TODO: PubKey
			suite,
			peeringProvider,
			registry,
			logger,
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
