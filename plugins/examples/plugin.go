package examples

import (
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfimpl"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/packages/vm/examples/sc7"
	"github.com/iotaledger/wasp/packages/vm/examples/sc8"
	"github.com/iotaledger/wasp/packages/vm/examples/sc9"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/vmnil"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const PluginName = "Examples"

type example struct {
	programHash  string
	getProcessor func() vmtypes.Processor
	name         string
}

func Init() *node.Plugin {
	return node.NewPlugin(PluginName, node.Enabled, configure, run)
}

func configure(ctx *node.Plugin) {
	allExamples := []example{
		{vmnil.ProgramHash, vmnil.GetProcessor, "vmnil"},
		{logsc.ProgramHash, logsc.GetProcessor, "logsc"},
		{inccounter.ProgramHash, inccounter.GetProcessor, "inccounter"},
		{fairroulette.ProgramHash, fairroulette.GetProcessor, "FairRoulette"},
		//{wasmhost.ProgramHash, wasmhost.GetProcessor, "wasmpoc"},
		{fairauction.ProgramHash, fairauction.GetProcessor, "FairAuction"},
		{tokenregistry.ProgramHash, tokenregistry.GetProcessor, "TokenRegistry"},
		{sc7.ProgramHash, sc7.GetProcessor, "sc7"},
		{sc8.ProgramHash, sc8.GetProcessor, "sc8"},
		{sc9.ProgramHash, sc9.GetProcessor, "sc9"},
		{dwfimpl.ProgramHash, dwfimpl.GetProcessor, "DonateWithFeedback"},
	}

	for _, ex := range allExamples {
		hash, _ := hashing.HashValueFromBase58(ex.programHash)
		registry.RegisterBuiltinProgramMetadata(&hash, ex.name+" (Built-in Smart Contract example)")
		processor.RegisterBuiltinProcessor(&hash, ex.getProcessor)
	}
}

func run(ctx *node.Plugin) {
}
