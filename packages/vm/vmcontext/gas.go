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
	vmctx.task.Log.Debugf("gas budget: %d", gasBudget)
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
	vmctx.gasBurnedTotal += g

	if vmctx.gasBurnedTotal+g > vmctx.chainInfo.GasLimits.MaxGasPerBlock {
		panic(vmexceptions.ErrBlockGasLimitExceeded)
	}

	if vmctx.gasBurned > vmctx.gasBudgetAdjusted {
		panic(vm.ErrGasBudgetExceeded)
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
