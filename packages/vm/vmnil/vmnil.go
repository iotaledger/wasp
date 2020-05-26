package vmnil

import (
	"github.com/iotaledger/wasp/packages/vm"
)

type vmnil struct {
}

func New() vm.Processor {
	return vmnil{}
}

func (v vmnil) Run(inputs *vm.VMContext) {
	panic("implement me")
}
