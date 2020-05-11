package main

import (
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/banner"
	"github.com/iotaledger/wasp/plugins/cli"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/gracefulshutdown"
	"github.com/iotaledger/wasp/plugins/logger"
	"github.com/iotaledger/wasp/plugins/webapi"
)

var PLUGINS = node.Plugins(
	banner.Plugin,
	config.Plugin,
	logger.Plugin,
	gracefulshutdown.Plugin,
	webapi.Plugin,
	cli.Plugin,
)

func main() {
	node.Run(
		PLUGINS,
	)
}
