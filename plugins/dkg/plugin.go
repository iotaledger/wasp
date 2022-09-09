// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

import (
	"github.com/iotaledger/hive.go/logger"
	hive_node "github.com/iotaledger/hive.go/node"
	dkg_pkg "github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/registry"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

const pluginName = "DKG"

var defaultNode *dkg_pkg.Node // A singleton.

// Init is an entry point for the plugin.
func Init() *hive_node.Plugin {
	configure := func(_ *hive_node.Plugin) {
		log := logger.NewLogger(pluginName)
		reg := registry.DefaultRegistry()
		peeringProvider := peering.DefaultNetworkProvider()
		var err error
		defaultNode, err = dkg_pkg.NewNode(
			reg.GetNodeIdentity(),
			peeringProvider,
			reg,
			log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelWarn)).Sugar(),
		)
		if err != nil {
			panic(xerrors.Errorf("failed to initialize the DKG node: %w", err))
		}
	}
	run := func(_ *hive_node.Plugin) {
		// Nothing to run here.
	}
	return hive_node.NewPlugin(pluginName, nil, hive_node.Enabled, configure, run)
}

// DefaultNode returns the default instance of the DKG Node Provider.
// It should be used to access all the DKG Node functions (not the DKG Initiator's).
func DefaultNode() *dkg_pkg.Node {
	return defaultNode
}
