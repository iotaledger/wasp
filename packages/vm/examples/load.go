package examples

import (
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
	"github.com/iotaledger/wasp/packages/vm/examples/wasmhost"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

var allProcessors = map[string]func() vmtypes.Processor{
	vmnil.ProgramHash:         vmnil.GetProcessor,
	logsc.ProgramHash:         logsc.GetProcessor,
	inccounter.ProgramHash:    inccounter.GetProcessor,
	fairroulette.ProgramHash:  fairroulette.GetProcessor,
	wasmhost.ProgramHash:      wasmhost.GetProcessor,
	fairauction.ProgramHash:   fairauction.GetProcessor,
	tokenregistry.ProgramHash: tokenregistry.GetProcessor,
	sc7.ProgramHash:           sc7.GetProcessor,
	sc8.ProgramHash:           sc8.GetProcessor,
	sc9.ProgramHash:           sc9.GetProcessor,
	dwfimpl.ProgramHash:       dwfimpl.GetProcessor,
}

func GetProcessor(progHashStr string) (vmtypes.Processor, bool) {
	getProcessor, ok := allProcessors[progHashStr]
	return getProcessor(), ok
}
