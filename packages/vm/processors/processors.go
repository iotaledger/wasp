package processors

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type vmProcessor struct {
	index     coretypes.Uint16
	processor vmtypes.Processor
}

type ChainProcessors struct {
	processors map[coretypes.Uint16]*vmProcessor
}

func New() *ChainProcessors {
	ret := &ChainProcessors{
		processors: make(map[coretypes.Uint16]*vmProcessor),
	}
	proc, err := FromBinaryCode("builtinvm", nil)
	if err != nil {
		panic(err)
	}
	// always contains bootup processor
	ret.processors[0xFFFF] = &vmProcessor{
		index:     0xFFFF,
		processor: proc,
	}
	return ret
}

func (cps *ChainProcessors) AddProcessor(index coretypes.Uint16, programCode []byte, vmtype string) error {
	if _, ok := cps.processors[index]; ok {
		return fmt.Errorf("smart contract processor with index %d already exists", index)
	}
	proc, err := FromBinaryCode(vmtype, programCode)
	if err != nil {
		return err
	}
	cps.processors[index] = &vmProcessor{
		index:     index,
		processor: proc,
	}
	return nil
}

func (cps *ChainProcessors) GetProcessor(index uint16) (vmtypes.Processor, bool) {
	proc, ok := cps.processors[(coretypes.Uint16)(index)]
	return proc.processor, ok
}

func (cps *ChainProcessors) ExistsProcessor(index uint16) bool {
	_, ok := cps.processors[(coretypes.Uint16)(index)]
	return ok
}
