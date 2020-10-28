// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package factory

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type factoryProcessor struct{}

type factoryEntryPoint func(ctx vmtypes.Sandbox, params ...interface{})

var Processor = factoryProcessor{}

func (v factoryProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	switch code {
	case 0:
		return (factoryEntryPoint)(initialize), true

	case 1:
		return (factoryEntryPoint)(newContract), true
	}
	return nil, false
}

func (v factoryProcessor) GetDescription() string {
	return "Factory processor"
}

func (ep factoryEntryPoint) Call(ctx vmtypes.Sandbox, params ...interface{}) interface{} {
	ep(ctx, params...)
	return nil
}

func (ep factoryEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

const (
	VarStateInitialized = "i"
	VarChainID          = "c"
	VarContractRegistry = "r"
)
