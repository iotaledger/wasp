package registry

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/iotaledger/hive.go/logger"
	hive_node "github.com/iotaledger/hive.go/node"
	registry_pkg "github.com/iotaledger/wasp/packages/registry"
	tcrypto_pkg "github.com/iotaledger/wasp/packages/tcrypto"
)

const pluginName = "Registry"

var (
	defaultRegistry *registry_pkg.Impl // A singleton.
)

// DefaultRegistry returns an initialized default registry.
func DefaultRegistry() *registry_pkg.Impl {
	return defaultRegistry
}

// Init is an entry point for the plugin.
func Init(suite tcrypto_pkg.Suite) *hive_node.Plugin {
	configure := func(_ *hive_node.Plugin) {
		defaultRegistry = registry_pkg.NewRegistry(suite, logger.NewLogger(pluginName))
	}
	run := func(_ *hive_node.Plugin) {
		// Nothing to run here.
	}
	return hive_node.NewPlugin(pluginName, hive_node.Enabled, configure, run)
}

// InitFlags configures the relevant CLI flags.
func InitFlags() {
	registry_pkg.InitFlags()
}
