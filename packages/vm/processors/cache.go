package processors

import (
	"fmt"
	"sync"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// Cache stores all initialized VMProcessor instances used by a single chain
type Cache struct {
	mutex      *sync.Mutex
	Config     *Config
	processors map[hashing.HashValue]iscp.VMProcessor
}

func MustNew(config *Config) *Cache {
	ret := &Cache{
		mutex:      &sync.Mutex{},
		Config:     config,
		processors: make(map[hashing.HashValue]iscp.VMProcessor),
	}
	// default builtin processor has root contract hash
	err := ret.NewProcessor(root.Contract.ProgramHash, nil, vmtypes.Core)
	if err != nil {
		panic(err)
	}
	return ret
}

// NewProcessor deploys new processor in the cache
func (cps *Cache) NewProcessor(programHash hashing.HashValue, programCode []byte, vmtype string) error {
	cps.mutex.Lock()
	defer cps.mutex.Unlock()

	return cps.newProcessor(programHash, programCode, vmtype)
}

func (cps *Cache) newProcessor(programHash hashing.HashValue, programCode []byte, vmtype string) error {
	var proc iscp.VMProcessor
	var ok bool
	var err error

	if cps.ExistsProcessor(programHash) {
		return nil
	}
	switch vmtype {
	case vmtypes.Core:
		proc, err = core.GetProcessor(programHash)
		if err != nil {
			return err
		}

	case vmtypes.Native:
		if proc, ok = cps.Config.GetNativeProcessor(programHash); !ok {
			return fmt.Errorf("NewProcessor: can't load example processor with hash %s", programHash.String())
		}

	default:
		proc, err = cps.Config.NewProcessorFromBinary(vmtype, programCode)
		if err != nil {
			return err
		}
	}
	cps.processors[programHash] = proc
	return nil
}

func (cps *Cache) ExistsProcessor(h hashing.HashValue) bool {
	_, ok := cps.processors[h]
	return ok
}

func (cps *Cache) GetOrCreateProcessor(rec *root.ContractRecord, getBinary func(hashing.HashValue) (string, []byte, error)) (iscp.VMProcessor, error) {
	return cps.GetOrCreateProcessorByProgramHash(rec.ProgramHash, getBinary)
}

func (cps *Cache) GetOrCreateProcessorByProgramHash(progHash hashing.HashValue, getBinary func(hashing.HashValue) (string, []byte, error)) (iscp.VMProcessor, error) {
	cps.mutex.Lock()
	defer cps.mutex.Unlock()

	if proc, ok := cps.processors[progHash]; ok {
		return proc, nil
	}
	vmtype, binary, err := getBinary(progHash)
	if err != nil {
		return nil, fmt.Errorf("internal error: can't get the binary for the program: %v", err)
	}
	if err := cps.newProcessor(progHash, binary, vmtype); err != nil {
		return nil, err
	}
	if proc, ok := cps.processors[progHash]; ok {
		return proc, nil
	}
	return nil, fmt.Errorf("internal error: can't get the deployed processor")
}

// RemoveProcessor deletes processor from cache
func (cps *Cache) RemoveProcessor(h hashing.HashValue) {
	cps.mutex.Lock()
	defer cps.mutex.Unlock()
	delete(cps.processors, h)
}
