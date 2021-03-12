package processors

import (
	"fmt"
	"github.com/iotaledger/wasp/contracts/native"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"sync"
)

// ProcessorCache is an object maintained by each chain
type ProcessorCache struct {
	*sync.Mutex
	processors map[hashing.HashValue]coretypes.Processor
}

func MustNew() *ProcessorCache {
	ret := &ProcessorCache{
		Mutex:      &sync.Mutex{},
		processors: make(map[hashing.HashValue]coretypes.Processor),
	}
	// default builtin processor has root contract hash
	err := ret.NewProcessor(root.Interface.ProgramHash, nil, core.VMType)
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
	var proc coretypes.Processor
	var ok bool
	var err error

	if cps.ExistsProcessor(programHash) {
		return nil
	}
	switch vmtype {
	case core.VMType:
		proc, err = core.GetProcessor(programHash)
		if err != nil {
			return err
		}

	case native.VMType:
		if proc, ok = native.GetProcessor(programHash); !ok {
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

func (cps *ProcessorCache) ExistsProcessor(h hashing.HashValue) bool {
	_, ok := cps.processors[h]
	return ok
}

func (cps *ProcessorCache) GetOrCreateProcessor(rec *root.ContractRecord, getBinary func(hashing.HashValue) (string, []byte, error)) (coretypes.Processor, error) {
	return cps.GetOrCreateProcessorByProgramHash(rec.ProgramHash, getBinary)
}

func (cps *ProcessorCache) GetOrCreateProcessorByProgramHash(progHash hashing.HashValue, getBinary func(hashing.HashValue) (string, []byte, error)) (coretypes.Processor, error) {
	cps.Lock()
	defer cps.Unlock()

	if proc, ok := cps.processors[progHash]; ok {
		return proc, nil
	}
	vmtype, binary, err := getBinary(progHash)
	if err != nil {
		return nil, fmt.Errorf("internal error: can't get the binary for the program: %v", err)
	}
	if err = cps.newProcessor(progHash, binary, vmtype); err != nil {
		return nil, err
	}
	if proc, ok := cps.processors[progHash]; ok {
		return proc, nil
	}
	return nil, fmt.Errorf("internal error: can't get the deployed processor")
}

// RemoveProcessor deletes processor from cache
func (cps *ProcessorCache) RemoveProcessor(h hashing.HashValue) {
	cps.Lock()
	defer cps.Unlock()
	delete(cps.processors, h)
}
