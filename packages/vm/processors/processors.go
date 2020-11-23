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
	_, err := ret.NewProcessor(root.ProgramHash[:], builtinvm.VMType)
	if err != nil {
		panic(err)
	}
	return ret
}

// NewProcessor deploys new processor in the cache or return existing
func (cps *ProcessorCache) NewProcessor(programCode []byte, vmtype string) (*hashing.HashValue, error) {
	cps.Lock()
	defer cps.Unlock()

	return cps.newProcessor(programCode, vmtype)
}

func (cps *ProcessorCache) newProcessor(programCode []byte, vmtype string) (*hashing.HashValue, error) {
	var proc vmtypes.Processor
	var err error
	var ok bool
	var programHash hashing.HashValue

	switch vmtype {
	case builtinvm.VMType:
		programHash, err = hashing.HashValueFromBytes(programCode)
		if err != nil {
			return nil, fmt.Errorf("NewProcessor: %v", err)
		}
		if cps.ExistsProcessor(&programHash) {
			return &programHash, nil
		}
		proc, err = builtinvm.GetProcessor(programHash)
		if err != nil {
			return nil, err
		}

	case examples.VMType:
		programHash, err = hashing.HashValueFromBytes(programCode)
		if err != nil {
			return nil, fmt.Errorf("NewProcessor: %v", err)
		}
		if cps.ExistsProcessor(&programHash) {
			return &programHash, nil
		}
		if proc, ok = examples.GetExampleProcessor(programHash.String()); !ok {
			return nil, fmt.Errorf("NewProcessor: can't load example processor with hash %s", programHash.String())
		}

	default:
		programHash = deploymentHash(programCode, vmtype)
		if cps.ExistsProcessor(&programHash) {
			return &programHash, nil
		}
		proc, err = NewProcessorFromBinary(vmtype, programCode)
		if err != nil {
			return nil, err
		}
	}
	cps.processors[programHash] = proc
	return &programHash, nil
}

func (cps *ProcessorCache) ExistsProcessor(h *hashing.HashValue) bool {
	_, ok := cps.processors[*h]
	return ok
}

func (cps *ProcessorCache) GetOrCreateProcessor(rec *root.ContractRecord, getBinary func(*hashing.HashValue) ([]byte, error)) (vmtypes.Processor, error) {
	cps.Lock()
	defer cps.Unlock()

	if proc, ok := cps.processors[rec.DeploymentHash]; ok {
		return proc, nil
	}
	binary, err := getBinary(&rec.DeploymentHash)
	if err != nil {
		return nil, fmt.Errorf("internal error: can't get the binary for the program: %v", err)
	}
	deploymentHash, err := cps.newProcessor(binary, rec.VMType)
	if err != nil {
		return nil, err
	}
	if *deploymentHash != rec.DeploymentHash {
		return nil, fmt.Errorf("internal error: *deploymentHash != deploymentHash")
	}
	if proc, ok := cps.processors[rec.DeploymentHash]; ok {
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

// deploymentHash helper function to calculate hash of the cache
func deploymentHash(programCode []byte, vmtype string) hashing.HashValue {
	return *hashing.HashData(programCode, []byte(vmtype))
}
