package main

import (
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/banner"
	"github.com/iotaledger/wasp/plugins/cli"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/dispatcher"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/iotaledger/wasp/plugins/logger"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/peering"
	roundtrip "github.com/iotaledger/wasp/plugins/testplugins/rountrip"
	"github.com/iotaledger/wasp/plugins/testplugins/scmeta"
	"github.com/iotaledger/wasp/plugins/webapi"
)

var PLUGINS = node.Plugins(
	banner.Plugin,
	config.Plugin,
	logger.Plugin,
	gracefulshutdown.Plugin,
	webapi.Plugin,
	cli.Plugin,
	database.Plugin,
	peering.Plugin,
	nodeconn.Plugin,
	dispatcher.Plugin,
)

var TestPLUGINS = node.Plugins(
	scmeta.Plugin,
	roundtrip.Plugin,
)

func main() {
	node.Run(
		PLUGINS,
		TestPLUGINS,
	)
}
