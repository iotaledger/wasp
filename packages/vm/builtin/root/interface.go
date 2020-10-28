// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package root

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type factoryProcessor struct{}

type factoryEntryPoint func(ctx vmtypes.Sandbox, params kv.ImmutableCodec) interface{}

var Processor = factoryProcessor{}

var (
	entryPointInitialize  = coretypes.NewEntryPointCodeFromFunctionName("initialize")
	entryPointNewContract = coretypes.NewEntryPointCodeFromFunctionName("newContract")
)

func (v factoryProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	switch code {
	case entryPointInitialize:
		return (factoryEntryPoint)(initialize), true

	case entryPointNewContract:
		return (factoryEntryPoint)(newContract), true
	}
	return nil, false
}

func (v factoryProcessor) GetDescription() string {
	return "Factory processor"
}

func (ep factoryEntryPoint) Call(ctx vmtypes.Sandbox, params kv.ImmutableCodec) interface{} {
	err := ep(ctx, params)
	if err != nil {
		if _, isError := err.(error); isError {
			ctx.Publishf("error occured: '%v'", err)
		}
	}
	return err
}

func (ep factoryEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

const (
	VarStateInitialized = "i"
	VarChainID          = "c"
	VarContractRegistry = "r"
)
