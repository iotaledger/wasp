package processor

import (
	"fmt"
	processors2 "github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"sync"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util/sema"
)

type processorInstance struct {
	vmtypes.Processor
	timedLock *sema.Lock
}

// TODO implement multiple workers/instances per program hash. Currently only one

var (
	processors        = make(map[string]processorInstance)
	processorsMutex   sync.RWMutex
	builtinProcessors = make(map[hashing.HashValue]func() vmtypes.Processor)
)

func RegisterBuiltinProcessor(programHash *hashing.HashValue, proc func() vmtypes.Processor) {
	builtinProcessors[*programHash] = proc
}

// LoadProcessorAsync creates and registers processor for program hash asynchronously
func LoadProcessorAsync(programHash *hashing.HashValue, onFinish func(err error)) {
	go func() {
		proc, err := loadProcessor(programHash)
		if err != nil {
			onFinish(err)
			return
		}

		processorsMutex.Lock()
		processors[programHash.String()] = processorInstance{
			Processor: proc,
			timedLock: sema.New(),
		}
		processorsMutex.Unlock()

		onFinish(nil)
	}()
}

// loadProcessor creates processor instance
func loadProcessor(progHash *hashing.HashValue) (vmtypes.Processor, error) {
	proc, ok := builtinProcessors[*progHash]
	if ok {
		return proc(), nil
	}

	md, err := registry.GetProgramMetadata(progHash)
	if err != nil {
		return nil, err
	}
	if md == nil {
		return nil, fmt.Errorf("Program metadata for hash %s not found", progHash.String())
	}

	if md.VMType == "builtin" {
		return nil, fmt.Errorf("Processor for builtin program %s not registered", progHash.String())
	}

	binaryCode, err := registry.GetProgramCode(progHash)
	if err != nil {
		return nil, err
	}

	return processors2.NewProcessorFromBinaryCode(md.VMType, binaryCode)
}
