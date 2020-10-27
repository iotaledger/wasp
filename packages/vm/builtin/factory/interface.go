// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package factory

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type bootupProcessor struct{}

type bootupEntryPoint func(ctx vmtypes.Sandbox)

var Processor = bootupProcessor{}

func (v bootupProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	switch code {
	case 0:
		return (bootupEntryPoint)(initialize), true

	case 1:
		return (bootupEntryPoint)(newContract), true
	}
	return nil, false
}

func (v bootupProcessor) GetDescription() string {
	return "Bootup processor"
}

func (ep bootupEntryPoint) Call(ctx vmtypes.Sandbox, params ...interface{}) interface{} {
	ep(ctx)
	return nil
}

func (ep bootupEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

const (
	VarStateInitialized = "i"
	VarChainID          = "c"
	VarContractRegistry = "r"
)
