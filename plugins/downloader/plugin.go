package downloader

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/downloader"
)

// PluginName is the name of the web API plugin.
const PluginName = "Downloader"

var (
	log *logger.Logger // Logger for Downloader functionality
)

func Init() *node.Plugin {
	var configure, run func(*node.Plugin)
	configure = func(*node.Plugin) {
		var log *logger.Logger = logger.NewLogger(PluginName)
		downloader.Init(log)
	}
	run = func(*node.Plugin) {
		// Nothing to run here
	}

	Plugin := node.NewPlugin(PluginName, node.Enabled, configure, run)
	return Plugin
}
