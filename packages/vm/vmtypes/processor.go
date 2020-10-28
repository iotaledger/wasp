// package present processor interface. It must be implemented by VM
package vmtypes

import "github.com/iotaledger/wasp/packages/coretypes"

// Processor is a abstract interface to the VM processor instance.
type Processor interface {
	GetEntryPoint(code coretypes.EntryPointCode) (EntryPoint, bool)
	GetDescription() string
}

// EntryPoint is an abstract interface by which VM is called by passing
// the Sandbox interface and optional parameters to it
// VM is expected to be fully deterministic and it result is 100% reflected
// as a side effect on the Sandbox interface
type EntryPoint interface {
	WithGasLimit(int) EntryPoint
	Call(ctx Sandbox, params ...interface{}) interface{}
}
