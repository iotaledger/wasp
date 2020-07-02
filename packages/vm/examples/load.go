package examples

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/examples/increasecounter"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/packages/vm/examples/vmnil"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func LoadProcessor(progHashStr string) (vmtypes.Processor, bool) {
	switch progHashStr {
	case vmnil.ProgramHash:
		return vmnil.GetProcessor(), true

	case logsc.ProgramHash:
		return logsc.GetProcessor(), true

	case fairroulette.ProgramHash:
		return fairroulette.GetProcessor(), true

	case increasecounter.ProgramHash:
		return increasecounter.GetProcessor(), true
	}
	return nil, false
}
