package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
)

func (vmctx *VMContext) EstimateRequiredStorageDeposit(par iscp.RequestParameters) uint64 {
	par.AdjustToMinimumDustDeposit = false
	out := transaction.BasicOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
	)
	return parameters.L1.Protocol.RentStructure.MinRent(out)
}
