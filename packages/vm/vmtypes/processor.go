// package present processor interface. It must be implemented by VM
package vmtypes

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
)

// Processor is a abstract interface to the VM processor instance.
type Processor interface {
	GetEntryPoint(code coretypes.EntryPointCode) (EntryPoint, bool)
	GetDescription() string
}

// EntryPoint is an abstract interface by which VM is called by passing
// the Sandbox interface and parameters to it
// the call from the request transaction has request argument as parameters
// the call from another contract can have any kv.Map
type EntryPoint interface {
	WithGasLimit(int) EntryPoint
	Call(ctx Sandbox, params kv.RCodec) interface{}
}
