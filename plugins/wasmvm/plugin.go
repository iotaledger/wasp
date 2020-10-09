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
	ctx.Node.Logger.Info("Configure " + PluginName)
	vmtypes.RegisterVMType(PluginName, wasmhost.GetProcessor)
	ctx.Node.Logger.Info("Configure " + PluginName + " done")
}

func run(ctx *node.Plugin) {
	ctx.Node.Logger.Info("Run " + PluginName)
}
