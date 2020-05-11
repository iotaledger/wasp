package main

import (
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/banner"
	"github.com/iotaledger/wasp/plugins/config"
	"github.com/iotaledger/wasp/plugins/logger"
)

var PLUGINS = node.Plugins(
	banner.Plugin,
	config.Plugin,
	logger.Plugin,
)

func main() {
	node.Run(
		PLUGINS,
	)
}
