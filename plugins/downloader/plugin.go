package downloader

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/downloader"
	"github.com/iotaledger/wasp/packages/parameters"
)

// PluginName is the name of the web API plugin.
const PluginName = "Downloader"

func Init() *node.Plugin {
	var configure, run func(*node.Plugin)
	configure = func(*node.Plugin) {
		var log *logger.Logger = logger.NewLogger(PluginName)
		downloader.Init(log, parameters.GetString(parameters.IpfsGatewayAddress))
	}
	run = func(*node.Plugin) {
		// Nothing to run here
	}

	Plugin := node.NewPlugin(PluginName, node.Enabled, configure, run)
	return Plugin
}
