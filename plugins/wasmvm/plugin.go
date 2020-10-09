package wasmvm

import (
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

const PluginName = "WasmVM"

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(ctx *node.Plugin) {
	vmtypes.RegisterVMType(PluginName, wasmhost.GetProcessor)
}

func run(ctx *node.Plugin) {
}
