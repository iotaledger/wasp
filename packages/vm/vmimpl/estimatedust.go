package vmimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
)

func (reqctx *requestContext) estimateRequiredStorageDeposit(par isc.RequestParameters) uint64 {
	par.AdjustToMinimumStorageDeposit = false
	out := transaction.BasicOutputFromPostData(
		reqctx.vm.task.AnchorOutput.AliasID.ToAddress(),
		reqctx.CurrentContractHname(),
		par,
	)
	return parameters.L1().Protocol.RentStructure.MinRent(out)
}
