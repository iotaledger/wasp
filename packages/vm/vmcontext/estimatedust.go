package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/transaction"
)

func (vmctx *VMContext) EstimateRequiredDustDeposit(par iscp.RequestParameters) uint64 {
	par.AdjustToMinimumDustDeposit = false
	out := transaction.BasicOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.L1Params.RentStructure(),
	)
	return out.VBytes(vmctx.task.L1Params.RentStructure(), nil)
}
