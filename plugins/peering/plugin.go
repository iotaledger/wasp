package peering

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	peering_pkg "github.com/iotaledger/wasp/packages/peering"
	peering_tcp "github.com/iotaledger/wasp/packages/peering/tcp"
	"github.com/labstack/gommon/log"
)

const (
	pluginName = "Peering"
)

var (
	defaultNetworkProvider *peering_tcp.NetImpl // A singleton instance.
)

// Init is an entry point for this plugin.
func Init() *node.Plugin {
	return node.NewPlugin(pluginName, node.Enabled, configure, run)
}

// DefaultNetworkProvider returns the default network provider implementation.
func DefaultNetworkProvider() peering_pkg.NetworkProvider {
	return defaultNetworkProvider
}

func configure(_ *node.Plugin) {
	var err error
	defaultNetworkProvider, err = peering_tcp.NewNetworkProvider(
		parameters.GetString(parameters.PeeringMyNetId),
		parameters.GetInt(parameters.PeeringPort),
		logger.NewLogger(pluginName),
	)
	if err != nil {
		panic(err)
	}
	log.Infof(
		"--------------------------------- NetID is %s -----------------------------------",
		defaultNetworkProvider.Self().Location(),
	)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(
		"WaspPeering",
		defaultNetworkProvider.Run,
		parameters.PriorityPeering,
	)
	if err != nil {
		panic(err)
	}
}
