// Wasp can have several VM types. Each of them can be represented by separate plugin
// Plugin name serves as a VM type during dynamic loading of the binary.
// VM plugins can be enabled/disabled in the configuration of the node instance
// wasmtimevm plugin statically links VM implemented with Wasmtime to Wasp
// be registering wasmhost.GetProcessor as function
package wasmtimevm

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/plugins/processors"
)

// pluginName is the name of the plugin.
const pluginName = "WasmTimeVM"

var log *logger.Logger

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, nil, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(pluginName)

	// register VM type(s)
	err := processors.Config.RegisterVMType(vmtypes.WasmTime, func(binary []byte) (iscp.VMProcessor, error) {
		// TODO (via config?) pass non-default timeout for WasmTime processor like this:
		// WasmTimeout = 3 * time.Second
		return wasmhost.GetProcessor(binary, log)
	})
	if err != nil {
		log.Panicf("%v: %v", pluginName, err)
	}
	log.Infof("registered VM type: '%s'", vmtypes.WasmTime)
}

func run(_ *node.Plugin) {
}
