package processors

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"sync"
)

type vmProcessor struct {
	index     coretypes.Uint16
	processor vmtypes.Processor
}

type ChainProcessors struct {
	*sync.Mutex
	processors map[coretypes.Uint16]*vmProcessor
}

func New() *ChainProcessors {
	ret := &ChainProcessors{
		Mutex:      &sync.Mutex{},
		processors: make(map[coretypes.Uint16]*vmProcessor),
	}
	// get factory processor
	proc, err := NewProcessorFromBinaryCode("builtinvm", nil)
	if err != nil {
		panic(err)
	}
	// factory processor always exist at index 0
	ret.processors[0] = &vmProcessor{
		index:     0,
		processor: proc,
	}
	return ret
}

func (cps *ChainProcessors) AddProcessor(index coretypes.Uint16, programCode []byte, vmtype string) error {
	cps.Lock()
	defer cps.Unlock()

	if _, ok := cps.processors[index]; ok {
		return fmt.Errorf("smart contract processor with index %d already exists", index)
	}
	proc, err := NewProcessorFromBinaryCode(vmtype, programCode)
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
	cps.Lock()
	defer cps.Unlock()

	proc, ok := cps.processors[(coretypes.Uint16)(index)]
	return proc.processor, ok
}

func (cps *ChainProcessors) ExistsProcessor(index uint16) bool {
	cps.Lock()
	defer cps.Unlock()

	_, ok := cps.processors[(coretypes.Uint16)(index)]
	return ok
}
