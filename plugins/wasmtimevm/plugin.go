// Wasp can have several VM types. Each of them can be represented by separate plugin
// Plugin name serves as a VM type during dynamic loading of the binary.
// VM plugins can be enabled/disabled in the configuration of the node instance
// wasmtimevm plugin statically links VM implemented with Wasmtime to Wasp
// be registering wasmhost.GetProcessor as function
package wasmtimevm

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/wasmproc"
)

// VMType is the name of the plugin.
const VMType = "wasmtimevm"

var log *logger.Logger

func Init() *node.Plugin {
	return node.NewPlugin(VMType, node.Enabled, configure, run)
}

func configure(_ *node.Plugin) {
	log = logger.NewLogger(VMType)

	// register VM type(s)
	err := processors.RegisterVMType(VMType, func(binary []byte) (coretypes.VMProcessor, error) {
		//TODO (via config?) pass non-default timeout for WasmTime processor like this:
		// WasmTimeout = 3 * time.Second
		return wasmproc.GetProcessor(binary, log)
	})
	if err != nil {
		log.Panicf("%v: %v", VMType, err)
	}
	log.Infof("registered VM type: '%s'", VMType)
}

func run(_ *node.Plugin) {
}
