package evmimpl

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm"
)

// this must only be callable from webapi or directly by the ISC VM, not from contract execution
func cannotBeCalledFromContracts(ctx isc.Sandbox) {
	// don't charge gas for this verification
	ctx.Privileged().GasBurnEnable(false)
	defer ctx.Privileged().GasBurnEnable(true)
	caller := ctx.Caller()
	if caller != nil && caller.Kind() == isc.AgentIDKindContract {
		panic(vm.ErrIllegalCall)
	}
}
