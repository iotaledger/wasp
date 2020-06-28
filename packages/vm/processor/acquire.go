package processor

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"time"
)

// CheckProcessor checks if processor instance is available.
// It doesn't attempt to acquire it
func CheckProcessor(programHash string) bool {
	processorsMutex.RLock()
	defer processorsMutex.RUnlock()

	_, ok := processors[programHash]
	return ok
}

const processorAcquireTimeout = 2 * time.Second

// Acquire takes one processor from worker pool for this program hash
// In case one processor in the pool it just attempts to lock it for the call
func Acquire(programHash string) (vmtypes.Processor, error) {
	processorsMutex.RLock()

	ret, ok := processors[programHash]
	if !ok {
		defer processorsMutex.RUnlock()
		return nil, fmt.Errorf("no such processor: %v", programHash)
	}
	processorsMutex.Unlock()

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
	processorsMutex.Unlock()

	ret.timedLock.Release()
}
