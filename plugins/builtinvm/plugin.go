package builtinvm

import (
	"fmt"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtin/root"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
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
	err := processors.RegisterVMType(PluginName, func(binaryCode []byte) (vmtypes.Processor, error) {
		if len(binaryCode) == 0 {
			// bootup processor
			return root.Processor, nil
		}
		programHash, err := hashing.HashValueFromBytes(binaryCode)
		if err != nil {
			return nil, err
		}
		ret, ok := examples.GetExampleProcessor(programHash.String())
		if !ok {
			return nil, fmt.Errorf("can't load example processor with hash %s", programHash.String())
		}
		return ret, nil
	})
	if err != nil {
		log.Panicf("%v: %v", PluginName, err)
	}
	log.Infof("registered VM type: '%s'", PluginName)
}

func run(_ *node.Plugin) {
}
