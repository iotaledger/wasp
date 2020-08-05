package logger

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/plugins/config"
)

// PluginName is the name of the logger plugin.
const PluginName = "Logger"

func Init() *node.Plugin {
	Plugin := node.NewPlugin(PluginName, node.Enabled)

	Plugin.Events.Init.Attach(events.NewClosure(func(*node.Plugin) {
		if err := logger.InitGlobalLogger(config.Node); err != nil {
			panic(err)
		}
	}))

	return Plugin
}
