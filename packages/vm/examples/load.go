package examples

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/packages/vm/examples/sc7"
	"github.com/iotaledger/wasp/packages/vm/examples/sc8"
	"github.com/iotaledger/wasp/packages/vm/examples/sc9"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/vmnil"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func LoadProcessor(progHashStr string) (vmtypes.Processor, bool) {
	switch progHashStr {
	case vmnil.ProgramHash:
		return vmnil.GetProcessor(), true

	case logsc.ProgramHash:
		return logsc.GetProcessor(), true

	case inccounter.ProgramHash:
		return inccounter.GetProcessor(), true

	case fairroulette.ProgramHash:
		return fairroulette.GetProcessor(), true

	case wasmpoc.ProgramHash:
		return wasmpoc.GetProcessor(), true

	case fairauction.ProgramHash:
		return fairauction.GetProcessor(), true

	case tokenregistry.ProgramHash:
		return tokenregistry.GetProcessor(), true

	case sc7.ProgramHash:
		return sc7.GetProcessor(), true

	case sc8.ProgramHash:
		return sc8.GetProcessor(), true

	case sc9.ProgramHash:
		return sc9.GetProcessor(), true
	}
	return nil, false
}
