package main

import (
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/parameters"
	_ "github.com/iotaledger/wasp/packages/vm/sandbox"
	"github.com/iotaledger/wasp/plugins/banner"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/iotaledger/wasp/plugins/cli"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/dashboard"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/dispatcher"
	"github.com/iotaledger/wasp/plugins/dkg"
	"github.com/iotaledger/wasp/plugins/globals"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/iotaledger/wasp/plugins/ipfs"
	"github.com/iotaledger/wasp/plugins/logger"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/publisher"
	"github.com/iotaledger/wasp/plugins/registry"
	"github.com/iotaledger/wasp/plugins/testplugins/nodeping"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/iotaledger/wasp/plugins/webapi"
	"go.dedis.ch/kyber/v3/pairing"
)

func main() {
	suite := pairing.NewSuiteBn256() // TODO: [KP] Single suite should be used in all the places.

	registry.InitFlags()
	parameters.InitFlags()

	plugins := node.Plugins(
		banner.Init(),
		config.Init(),
		logger.Init(),
		gracefulshutdown.Init(),
		webapi.Init(),
		ipfs.Init(),
		cli.Init(),
		database.Init(),
		registry.Init(suite),
		peering.Init(suite),
		dkg.Init(suite),
		nodeconn.Init(),
		dispatcher.Init(),
		chains.Init(),
		publisher.Init(),
		dashboard.Init(),
		wasmtimevm.Init(),
		globals.Init(),
	)

	testPlugins := node.Plugins(
		nodeping.Init(),
	)

	node.Run(
		plugins,
		testPlugins,
	)
}
