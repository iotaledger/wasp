package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"golang.org/x/xerrors"
)

func (vmctx *VMContext) gasBurnEnable(enable bool) {
	vmctx.gasBurnEnabled = enable
}

func (vmctx *VMContext) gasSetBudget(gasBudget uint64) {
	vmctx.gasBudget = gasBudget
	vmctx.gasBurned = 0
}

func (vmctx *VMContext) GasBurn(gas uint64) {
	if !vmctx.gasBurnEnabled {
		return
	}
	vmctx.gasBurned += gas
	if vmctx.gasBurned > vmctx.gasBudget {
		panic(xerrors.Errorf("%w: burned total: %d, budget %d", coreutil.ErrorGasBudgetExceeded, vmctx.gasBurned, vmctx.gasBudget))
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
