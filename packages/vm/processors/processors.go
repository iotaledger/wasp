package processors

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/nilprocessor"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"sync"
)

type ChainProcessors struct {
	*sync.Mutex
	processors []vmtypes.Processor
}

func New() (*ChainProcessors, error) {
	ret := &ChainProcessors{
		Mutex:      &sync.Mutex{},
		processors: make([]vmtypes.Processor, 0, 20),
	}
	index, err := ret.AddProcessor(nilprocessor.ProgramHash[:], "builtinvm")
	if err != nil {
		return nil, err
	}
	if index != 0 {
		panic("assertion failed: index != 0")
	}
	return ret, nil
}

func (cps *ChainProcessors) AddProcessor(programCode []byte, vmtype string) (uint16, error) {
	cps.Lock()
	defer cps.Unlock()

	proc, err := NewProcessorFromBinary(vmtype, programCode)
	if err != nil {
		return 0, err
	}
	cps.processors = append(cps.processors, proc)
	return uint16(len(cps.processors) - 1), nil
}

func (cps *ChainProcessors) GetProcessor(index uint16) (vmtypes.Processor, bool) {
	cps.Lock()
	defer cps.Unlock()

	if int(index) >= len(cps.processors) {
		return nil, false
	}
	return cps.processors[index], true
}

func (cps *ChainProcessors) RemoveProcessor(index uint16) error {
	cps.Lock()
	defer cps.Unlock()
	if int(index) >= len(cps.processors) {
		return fmt.Errorf("wrong index")
	}
	cps.processors[index] = nilprocessor.GetProcessor()
	return nil
}

func (cps *ChainProcessors) ExistsProcessor(index uint16) bool {
	cps.Lock()
	defer cps.Unlock()

	return int(index) < len(cps.processors)
}
