// Wasp can have several VM types. Each of them can be represented by separate plugin
// Plugin name serves as a VM type during dynamic loading of the binary.
// VM plugins can be enabled/disabled in the configuration of the node instance
// wasmtimevm plugin statically links VM implemented with Wasmtime to Wasp
// be registering wasmhost.GetProcessor as function
package wasmtimevm

import (
	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/core/processors"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "WasmTimeVM",
			Configure: configure,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
)

func configure() error {

	// register VM type(s)
	err := processors.Config.RegisterVMType(vmtypes.WasmTime, func(binary []byte) (isc.VMProcessor, error) {
		// TODO (via config?) pass non-default timeout for WasmTime processor like this:
		// WasmTimeout = 3 * time.Second
		return wasmhost.GetProcessor(binary, CoreComponent.Logger())
	})
	if err != nil {
		CoreComponent.LogPanic(err)
	}
	CoreComponent.LogInfof("registered VM type: '%s'", vmtypes.WasmTime)

	return nil
}
