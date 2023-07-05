package vmimpl

import (
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

func (vmctx *vmContext) GasBurnEnable(enable bool) {
	if enable && !vmctx.shouldChargeGasFee() {
		return
	}
	vmctx.blockGas.burnEnabled = enable
}

func (vmctx *vmContext) gasSetBudget(gasBudget, maxTokensToSpendForGasFee uint64) {
	vmctx.reqCtx.gas.budgetAdjusted = gasBudget
	vmctx.reqCtx.gas.maxTokensToSpendForGasFee = maxTokensToSpendForGasFee
	vmctx.reqCtx.gas.burned = 0
}

func (vmctx *vmContext) GasBurn(burnCode gas.BurnCode, par ...uint64) {
	if !vmctx.blockGas.burnEnabled {
		return
	}
	g := burnCode.Cost(par...)
	vmctx.reqCtx.gas.burnLog.Record(burnCode, g)
	vmctx.reqCtx.gas.burned += g

	if vmctx.reqCtx.gas.burned > vmctx.reqCtx.gas.budgetAdjusted {
		vmctx.reqCtx.gas.burned = vmctx.reqCtx.gas.budgetAdjusted // do not charge more than the limit set by the request
		panic(vm.ErrGasBudgetExceeded)
	}

	if vmctx.blockGas.burned+vmctx.reqCtx.gas.burned > vmctx.chainInfo.GasLimits.MaxGasPerBlock {
		panic(vmexceptions.ErrBlockGasLimitExceeded) // panic if the current request gas overshoots the block limit
	}
}

func (vmctx *vmContext) GasBudgetLeft() uint64 {
	if vmctx.reqCtx.gas.budgetAdjusted < vmctx.reqCtx.gas.burned {
		return 0
	}
	return vmctx.reqCtx.gas.budgetAdjusted - vmctx.reqCtx.gas.burned
}

func (vmctx *vmContext) GasBurned() uint64 {
	return vmctx.reqCtx.gas.burned
}