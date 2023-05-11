package vmcontext

import (
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

func (vmctx *VMContext) GasBurnEnable(enable bool) {
	if enable && !vmctx.shouldChargeGasFee() {
		return
	}
	vmctx.gasBurnEnabled = enable
}

func (vmctx *VMContext) gasSetBudget(gasBudget uint64) {
	vmctx.gasBudgetAdjusted = gasBudget
	vmctx.gasBurned = 0
}

func (vmctx *VMContext) GasBurn(burnCode gas.BurnCode, par ...uint64) {
	if !vmctx.gasBurnEnabled {
		return
	}
	g := burnCode.Cost(par...)
	vmctx.gasBurnLog.Record(burnCode, g)
	vmctx.gasBurned += g

	if vmctx.gasBurned > vmctx.gasBudgetAdjusted {
		vmctx.gasBurned = vmctx.gasBudgetAdjusted // do not charge more than the limit set by the request
		panic(vm.ErrGasBudgetExceeded)
	}

	if vmctx.gasBurnedTotal+vmctx.gasBurned > vmctx.chainInfo.GasLimits.MaxGasPerBlock {
		panic(vmexceptions.ErrBlockGasLimitExceeded) // panic if the current request gas overshoots the block limit
	}
}

func (vmctx *VMContext) GasBudgetLeft() uint64 {
	if vmctx.gasBudgetAdjusted < vmctx.gasBurned {
		return 0
	}
	return vmctx.gasBudgetAdjusted - vmctx.gasBurned
}

func (vmctx *VMContext) GasBurned() uint64 {
	return vmctx.gasBurned
}
