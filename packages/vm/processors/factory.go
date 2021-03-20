package processors

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"sync"
)

type VMConstructor func(binaryCode []byte) (coretypes.VMProcessor, error)

var (
	vmconstructors = make(map[string]VMConstructor)
	vmfactoryMutex sync.Mutex
)

// RegisterVMType registers new VM type by providing a constructor function to construct
// an instance of the processor.
// The constructor is a closure which also may encompass configuration params for the VM
// The function is normally called from the init code
func RegisterVMType(vmtype string, constructor VMConstructor) error {
	vmfactoryMutex.Lock()
	defer vmfactoryMutex.Unlock()

	if _, ok := vmconstructors[vmtype]; ok {
		return fmt.Errorf("duplicate vm type '%s'", vmtype)
	}
	vmconstructors[vmtype] = constructor
	return nil
}

// NewProcessorFromBinary creates an instance of the processor by its VM type and the binary code
func NewProcessorFromBinary(vmtype string, binaryCode []byte) (coretypes.VMProcessor, error) {
	vmfactoryMutex.Lock()
	defer vmfactoryMutex.Unlock()

	constructor, ok := vmconstructors[vmtype]
	if !ok {
		return nil, fmt.Errorf("unknown VM type '%s'", vmtype)
	}
	return constructor(binaryCode)
}
