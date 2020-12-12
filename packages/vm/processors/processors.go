package processors

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/builtinvm"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"sync"
)

// ProcessorCache is an object maintained by each chain
type ProcessorCache struct {
	*sync.Mutex
	processors map[hashing.HashValue]vmtypes.Processor
}

func MustNew() *ProcessorCache {
	ret := &ProcessorCache{
		Mutex:      &sync.Mutex{},
		processors: make(map[hashing.HashValue]vmtypes.Processor),
	}
	// default builtin processor has root contract hash
	err := ret.NewProcessor(root.Interface.ProgramHash, nil, builtinvm.VMType)
	if err != nil {
		panic(err)
	}
	return ret
}

// NewProcessor deploys new processor in the cache
func (cps *ProcessorCache) NewProcessor(programHash hashing.HashValue, programCode []byte, vmtype string) error {
	cps.Lock()
	defer cps.Unlock()

	return cps.newProcessor(programHash, programCode, vmtype)
}

func (cps *ProcessorCache) newProcessor(programHash hashing.HashValue, programCode []byte, vmtype string) error {
	var proc vmtypes.Processor
	var ok bool
	var err error

	if cps.ExistsProcessor(&programHash) {
		return nil
	}
	switch vmtype {
	case builtinvm.VMType:
		proc, err = builtinvm.GetProcessor(programHash)
		if err != nil {
			return err
		}

	case examples.VMType:
		if proc, ok = examples.GetExampleProcessor(programHash); !ok {
			return fmt.Errorf("NewProcessor: can't load example processor with hash %s", programHash.String())
		}

	default:
		proc, err = NewProcessorFromBinary(vmtype, programCode)
		if err != nil {
			return err
		}
	}
	cps.processors[programHash] = proc
	return nil
}

func (cps *ProcessorCache) ExistsProcessor(h *hashing.HashValue) bool {
	_, ok := cps.processors[*h]
	return ok
}

func (cps *ProcessorCache) GetOrCreateProcessor(rec *root.ContractRecord, getBinary func(hashing.HashValue) (string, []byte, error)) (vmtypes.Processor, error) {
	cps.Lock()
	defer cps.Unlock()

	if proc, ok := cps.processors[rec.ProgramHash]; ok {
		return proc, nil
	}
	vmtype, binary, err := getBinary(rec.ProgramHash)
	if err != nil {
		return nil, fmt.Errorf("internal error: can't get the binary for the program: %v", err)
	}
	if err = cps.newProcessor(rec.ProgramHash, binary, vmtype); err != nil {
		return nil, err
	}
	if proc, ok := cps.processors[rec.ProgramHash]; ok {
		return proc, nil
	}
	return nil, fmt.Errorf("internal error: can't get the deployed processor")
}

// RemoveProcessor deletes processor from cache
func (cps *ProcessorCache) RemoveProcessor(h *hashing.HashValue) {
	cps.Lock()
	defer cps.Unlock()
	delete(cps.processors, *h)
}
