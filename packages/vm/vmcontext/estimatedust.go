package vmcontext

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
)

func (vmctx *VMContext) EstimateRequiredStorageDeposit(par isc.RequestParameters) uint64 {
	par.AdjustToMinimumStorageDeposit = false
	out := transaction.BasicOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
	)
	return parameters.L1().Protocol.RentStructure.MinRent(out)
}
