package evmimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
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

func restricted(handler coreutil.Handler) coreutil.Handler {
	return func(ctx isc.Sandbox) dict.Dict {
		cannotBeCalledFromContracts(ctx)
		return handler(ctx)
	}
}
