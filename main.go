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
	databaseplugin "github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/dkg"
	"github.com/iotaledger/wasp/plugins/downloader"
	"github.com/iotaledger/wasp/plugins/globals"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/iotaledger/wasp/plugins/logger"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
	"github.com/iotaledger/wasp/plugins/publishernano"
	"github.com/iotaledger/wasp/plugins/registry"
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
		downloader.Init(),
		cli.Init(),
		databaseplugin.Init(),
		registry.Init(suite),
		peering.Init(suite),
		dkg.Init(suite),
		nodeconn.Init(),
		chains.Init(),
		publishernano.Init(),
		dashboard.Init(),
		wasmtimevm.Init(),
		globals.Init(),
	)

	node.Run(
		plugins,
	)
}
