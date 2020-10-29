package builtinvm

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// PluginName is the name of the plugin.
const PluginName = "builtinvm"

var log *logger.Logger

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)

	// treats binary as program hash
	err := processors.RegisterVMType(PluginName, builtinvm.Constructor)
	if err != nil {
		log.Panicf("%v: %v", PluginName, err)
	}
	log.Infof("registered VM type: '%s'", PluginName)
}

func run(_ *node.Plugin) {
}
