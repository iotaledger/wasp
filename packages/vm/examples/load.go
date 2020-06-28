package examples

import (
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/packages/vm/examples/vmnil"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func LoadProcessor(progHashStr string) (vmtypes.Processor, bool) {
	switch progHashStr {
	case vmnil.ProgramHash:
		return vmnil.New(), true

	case logsc.ProgramHash:
		return logsc.New(), true
	}
	return nil, false
}
