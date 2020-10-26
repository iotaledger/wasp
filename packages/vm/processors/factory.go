package processors

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"sync"
)

type VMConstructor func(binaryCode []byte) (vmtypes.Processor, error)

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

// FromBinaryCode creates an instance of the processor by its VM type and the binary code
func FromBinaryCode(vmtype string, binaryCode []byte) (vmtypes.Processor, error) {
	vmfactoryMutex.Lock()
	defer vmfactoryMutex.Unlock()

	constructor, ok := vmconstructors[vmtype]
	if !ok {
		return nil, fmt.Errorf("unknown VM type '%s'", vmtype)
	}
	return constructor(binaryCode)
}
