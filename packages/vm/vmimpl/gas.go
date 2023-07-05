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
	vmctx.reqctx.gas.budgetAdjusted = gasBudget
	vmctx.reqctx.gas.maxTokensToSpendForGasFee = maxTokensToSpendForGasFee
	vmctx.reqctx.gas.burned = 0
}

func (vmctx *vmContext) GasBurn(burnCode gas.BurnCode, par ...uint64) {
	if !vmctx.blockGas.burnEnabled {
		return
	}
	g := burnCode.Cost(par...)
	vmctx.reqctx.gas.burnLog.Record(burnCode, g)
	vmctx.reqctx.gas.burned += g

	if vmctx.reqctx.gas.burned > vmctx.reqctx.gas.budgetAdjusted {
		vmctx.reqctx.gas.burned = vmctx.reqctx.gas.budgetAdjusted // do not charge more than the limit set by the request
		panic(vm.ErrGasBudgetExceeded)
	}

	if vmctx.blockGas.burned+vmctx.reqctx.gas.burned > vmctx.chainInfo.GasLimits.MaxGasPerBlock {
		panic(vmexceptions.ErrBlockGasLimitExceeded) // panic if the current request gas overshoots the block limit
	}
}

func (vmctx *vmContext) GasBudgetLeft() uint64 {
	if vmctx.reqctx.gas.budgetAdjusted < vmctx.reqctx.gas.burned {
		return 0
	}
	return vmctx.reqctx.gas.budgetAdjusted - vmctx.reqctx.gas.burned
}

func (vmctx *vmContext) GasBurned() uint64 {
	return vmctx.reqctx.gas.burned
}
