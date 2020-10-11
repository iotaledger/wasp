package initvm

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// PluginName is the name of the RunVM plugin.
const PluginName = "RunVM"

var log *logger.Logger

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)

	// register VM type(s)
	err := vmtypes.RegisterVMType(wasmhost.VMType, wasmhost.GetProcessor)
	if err != nil {
		log.Panicf("RunVM: %v", err)
	}
	log.Infof("registered VM type: '%s'", wasmhost.VMType)
}

func run(_ *node.Plugin) {
}
