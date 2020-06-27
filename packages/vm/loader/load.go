package loader

import (
	"fmt"
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
		defer processorsMutex.Unlock()

		processors[programHash] = proc
		onFinish(nil)
	}()
}

func loadProcessor(programHash string) (processor.Processor, error) {
	switch programHash {
	case vmnil.ProgramHash:
		return vmnil.New(), nil

	case logsc.ProgramHash:
		return logsc.New(), nil

	default:
		return nil, fmt.Errorf("can't create processor for program hash %s", programHash)
	}
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
