package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/transaction"
)

func (vmctx *VMContext) EstimateRequiredDustDeposit(par iscp.RequestParameters) uint64 {
	par.AdjustToMinimumDustDeposit = false
	out := transaction.OutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.RentStructure,
	)
	return out.VByteCost(vmctx.task.RentStructure, nil)
}
