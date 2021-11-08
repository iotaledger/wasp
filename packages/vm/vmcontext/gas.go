package vmcontext

import "github.com/iotaledger/wasp/packages/iscp/coreutil"

func (vmctx *VMContext) GasSetBudget(gas int64) {
	vmctx.gas = gas
	vmctx.gasBudgetEnabled = vmctx.gas > 0
}

func (vmctx *VMContext) GasBurn(gas int64) {
	vmctx.gas -= gas
	if vmctx.gasBudgetEnabled && vmctx.gas < 0 {
		panic(coreutil.ErrorGasBudgetExceeded)
	}
}

func (vmctx *VMContext) GasBudget() int64 {
	return vmctx.gas
}
