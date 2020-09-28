package processor

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// CheckProcessor checks if processor instance is available.
// It doesn't attempt to acquire it
func CheckProcessor(programHash string) bool {
	processorsMutex.RLock()
	defer processorsMutex.RUnlock()

	_, ok := processors[programHash]
	return ok
}

// Acquire takes one processor from worker pool for this program hash
// In case only one processor is in the pool it just attempts to lock it for the call
// TODO implement pool of worker-processors
func Acquire(programHash string) (vmtypes.Processor, error) {
	processorsMutex.RLock()

	ret, ok := processors[programHash]
	if !ok {
		defer processorsMutex.RUnlock()
		return nil, fmt.Errorf("no such processor: %v", programHash)
	}
	processorsMutex.RUnlock()

	if !ret.timedLock.Acquire(processorAcquireTimeout) {
		return nil, fmt.Errorf("timeout: wasn't able to acquire processor for %v", processorAcquireTimeout)
	}
	return ret, nil
}

// Release releases processor for subsequent calls
func Release(programHash string) {
	processorsMutex.RLock()

	ret, ok := processors[programHash]
	if !ok {
		processorsMutex.RUnlock()
		return
	}
	processorsMutex.RUnlock()

	ret.timedLock.Release()
}
