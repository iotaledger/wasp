package examples

import (
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/packages/vm/examples/vmnil"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const VMType = "examplevm"

var allExamples = map[string]vmtypes.Processor{
	vmnil.ProgramHash:         vmnil.GetProcessor(),
	logsc.ProgramHash:         logsc.GetProcessor(),
	inccounter.ProgramHashStr: inccounter.GetProcessor(),
	// TODO
	//fairroulette.ProgramHashStr:  fairroulette.GetProcessor(),
	//fairauction.ProgramHashStr:   fairauction.GetProcessor(),
	//tokenregistry.ProgramHashStr: tokenregistry.GetProcessor(),
	//dwfimpl.ProgramHashStr:       dwfimpl.GetProcessor(),
}

func GetExampleProcessor(progHash string) (vmtypes.Processor, bool) {
	ret, ok := allExamples[progHash]
	if !ok {
		return nil, false
	}
	return ret, true
}
