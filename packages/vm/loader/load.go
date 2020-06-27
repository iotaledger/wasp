package loader

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/packages/vm/examples/vmnil"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"sync"
)

var (
	processors      = make(map[string]processor.Processor)
	processorsMutex sync.RWMutex
)

// LoadProcessorAsync creates and registers processor for program hash
// asynchronously
// possibly, locates Wasm program code in IPFS and caches here
func LoadProcessorAsync(programHash string, onFinish func(err error)) {
	go func() {
		proc, err := loadProcessor(programHash)
		if err != nil {
			onFinish(err)
			return
		}

		processorsMutex.Lock()
		processors[programHash] = proc
		processorsMutex.Unlock()

		onFinish(nil)
	}()
}

func loadProcessor(progHashStr string) (processor.Processor, error) {
	switch progHashStr {
	case vmnil.ProgramHash:
		return vmnil.New(), nil

	case logsc.ProgramHash:
		return logsc.New(), nil

	default:
		progHash, err := hashing.HashValueFromBase58(progHashStr)
		binaryCode, exist, err := registry.GetProgramCode(&progHash)

		if err != nil {
			return nil, err
		}
		if exist {
			return processorFromBinaryCode(binaryCode)
		}

		md, exist, err := registry.GetProgramMetadata(&progHash)
		if err != nil {
			return nil, err
		}
		if exist {
			binaryCode, err := loadBinaryCode(md.Location, &progHash)
			if err != nil {
				return nil, fmt.Errorf("failed to load program's binary data from location %s, program hash = %s",
					md.Location, progHashStr)
			}
			return processorFromBinaryCode(binaryCode)
		}
		return nil, fmt.Errorf("no metadata for program hash %s", progHashStr)
	}
}

func processorFromBinaryCode(binaryCode []byte) (processor.Processor, error) {
	panic("implement me")
}

// loads binary code of the VM, possibly from remote location
// caches it into the the registry
func loadBinaryCode(location string, progHash *hashing.HashValue) ([]byte, error) {
	panic("implement me")
}

func CheckProcessor(programHash string) bool {
	_, err := GetProcessor(programHash)
	return err == nil
}

func GetProcessor(programHash string) (processor.Processor, error) {
	processorsMutex.RLock()
	defer processorsMutex.RUnlock()

	ret, ok := processors[programHash]
	if !ok {
		return nil, fmt.Errorf("no such processor: %v", programHash)
	}
	return ret, nil
}
