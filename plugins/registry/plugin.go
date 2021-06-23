// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/iotaledger/hive.go/logger"
	hive_node "github.com/iotaledger/hive.go/node"
	registry "github.com/iotaledger/wasp/packages/registry"
)

const pluginName = "Registry"

var (
	defaultRegistry *registry.Impl // A singleton.
)

// DefaultRegistry returns an initialized default registry.
func DefaultRegistry() *registry.Impl {
	return defaultRegistry
}

// Init is an entry point for the plugin.
func Init() *hive_node.Plugin {
	configure := func(_ *hive_node.Plugin) {
		defaultRegistry = registry.NewRegistry(logger.NewLogger(pluginName), nil)
	}
	run := func(_ *hive_node.Plugin) {
		// Nothing to run here.
	}
	return hive_node.NewPlugin(pluginName, hive_node.Enabled, configure, run)
}

// InitFlags configures the relevant CLI flags.
func InitFlags() {
	registry.InitFlags()
}
