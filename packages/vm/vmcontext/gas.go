package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"golang.org/x/xerrors"
)

func (vmctx *VMContext) gasBurnEnable(enable bool) {
	vmctx.gasBurnEnabled = enable
}

func (vmctx *VMContext) gasSetBudget(gasBudget uint64) {
	vmctx.gasBudget = gasBudget
	vmctx.gasBurned = 0
}

func (vmctx *VMContext) GasBurn(g uint64, burnCode gas.BurnCode) {
	if !vmctx.gasBurnEnabled {
		return
	}
	vmctx.gasBurnLog.Record(burnCode, g)
	vmctx.gasBurned += g
	if vmctx.gasBurned > vmctx.gasBudget {
		panic(xerrors.Errorf("%v: burned (budget) = %d (%d)",
			coreutil.ErrorGasBudgetExceeded, vmctx.gasBurned, vmctx.gasBudget))
	}
}

func (vmctx *VMContext) GasBudgetLeft() uint64 {
	if vmctx.gasBudget < vmctx.gasBurned {
		return 0
	}
	return vmctx.gasBudget - vmctx.gasBurned
}

func (vmctx *VMContext) GasBurned() uint64 {
	return vmctx.gasBurned
}
