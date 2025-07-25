package execution

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

// this file holds functions common to both context implementation (viewcontext and vmcontext)

func GetEntryPoint(ctx WaspContext, targetContract, epCode isc.Hname) isc.VMProcessorEntryPoint {
	proc, ok := ctx.Processors().GetCoreProcessor(targetContract)
	if !ok {
		if gasctx, ok2 := ctx.(GasContext); ok2 {
			gasctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		}
		panic(vm.ErrContractNotFound.Create(uint32(targetContract)))
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		if gasctx, ok2 := ctx.(GasContext); ok2 {
			gasctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		}
		panic(vm.ErrTargetEntryPointNotFound)
	}
	return ep
}
