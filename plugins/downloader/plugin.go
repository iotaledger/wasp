package downloader

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/downloader"
	"github.com/iotaledger/wasp/packages/parameters"
)

// PluginName is the name of the downloader plugin.
const PluginName = "Downloader"

//Init inits the plugin.
func Init() *node.Plugin {
	configure := func(*node.Plugin) {
		log := logger.NewLogger(PluginName)
		downloader.Init(log, parameters.GetString(parameters.IpfsGatewayAddress))
	}
	run := func(*node.Plugin) {
		// Nothing to run here
	}

	Plugin := node.NewPlugin(PluginName, node.Enabled, configure, run)
	return Plugin
}
