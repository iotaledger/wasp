package vmimpl

import (
	"github.com/ethereum/go-ethereum/common"
	
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

func (reqctx *requestContext) estimateRequiredStorageDeposit(par isc.RequestParameters) coin.Value {
	par.AdjustToMinimumStorageDeposit = false

	hname := reqctx.CurrentContractHname()
	contractIdentity := isc.ContractIdentityFromHname(hname)
	if hname == evm.Contract.Hname() {
		contractIdentity = isc.ContractIdentityFromEVMAddress(common.Address{}) // use empty EVM address as STUB
	}

	panic("refactor me: transaction.BasicOutputFromPostData")
	_ = contractIdentity

	//return parameters.L1().Protocol.RentStructure.MinRent(out)
	return 0
}
