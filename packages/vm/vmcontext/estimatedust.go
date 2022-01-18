package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/transaction"
)

// TODO missing gas burn
func (vmctx *VMContext) EstimateRequiredDustDeposit(par iscp.RequestParameters) (uint64, error) {
	par.AdjustToMinimumDustDeposit = false
	out := transaction.ExtendedOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.RentStructure,
	)
	return out.VByteCost(vmctx.task.RentStructure, nil), nil
}
