package vmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

func (reqctx *requestContext) estimateRequiredStorageDeposit(par isc.RequestParameters) uint64 {
	par.AdjustToMinimumStorageDeposit = false

	hname := reqctx.CurrentContractHname()
	contractIdentity := isc.ContractIdentityFromHname(hname)
	if hname == evm.Contract.Hname() {
		contractIdentity = isc.ContractIdentityFromEVMAddress(common.Address{}) // use empty EVM address as STUB
	}

	panic("refactor me: transaction.BasicOutputFromPostData")
	_ = contractIdentity
	var out iotago.Output

	return parameters.L1().Protocol.RentStructure.MinRent(out)
}
