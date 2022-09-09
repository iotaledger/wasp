package logger

import (
	"github.com/iotaledger/hive.go/configuration"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"go.uber.org/dig"
)

// PluginName is the name of the logger plugin.
const PluginName = "Logger"

func Init(conf *configuration.Configuration) *node.Plugin {
	Plugin := node.NewPlugin(PluginName, nil, node.Enabled)

	Plugin.Events.Init.Attach(events.NewClosure(func(*node.Plugin, *dig.Container) {
		if err := logger.InitGlobalLogger(conf); err != nil {
			panic(err)
		}
		initGoEthLogger(logger.NewLogger("go-ethereum"))
	}))

	return Plugin
}
