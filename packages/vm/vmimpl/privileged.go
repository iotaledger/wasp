package vmimpl

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

func (reqctx *requestContext) mustBeCalledFromContract(contract *coreutil.ContractInfo) {
	if reqctx.CurrentContractHname() != contract.Hname() {
		panic(fmt.Sprintf("%v: core contract '%s' expected", vm.ErrPrivilegedCallFailed, contract.Name))
	}
}

func (reqctx *requestContext) CreateNewFoundry(scheme isc.SimpleTokenScheme, metadata []byte) (uint32, uint64) {
	reqctx.mustBeCalledFromContract(accounts.Contract)
	panic("CreateNewFoundry not implemented yet")
	// return reqctx.vm.txbuilder.CreateNewFoundry(scheme, metadata)
}

func (reqctx *requestContext) DestroyFoundry(sn uint32) uint64 {
	reqctx.mustBeCalledFromContract(accounts.Contract)
	panic("DestroyFoundry not implemented yet")
	// return reqctx.vm.txbuilder.DestroyFoundry(sn)
}

func (reqctx *requestContext) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	reqctx.mustBeCalledFromContract(accounts.Contract)
	panic("ModifyFoundrySupply not implemented yet")
	// out, _ := reqctx.accountsStateWriter(false).GetFoundryOutput(sn, reqctx.ChainID())
	// nativeTokenID, err := out.NativeTokenID()
	// if err != nil {
	// 	panic(fmt.Errorf("internal: %w", err))
	// }
	// return reqctx.vm.txbuilder.ModifyNativeTokenSupply(nativeTokenID, delta)
}

func (reqctx *requestContext) CallOnBehalfOf(caller isc.AgentID, msg isc.Message, allowance *isc.Assets) isc.CallArguments {
	reqctx.Debugf("CallOnBehalfOf: caller = %s, msg = %s", caller.String(), msg)
	return reqctx.callProgram(msg, allowance, caller)
}

func (reqctx *requestContext) SendOnBehalfOf(caller isc.ContractIdentity, metadata isc.RequestParameters) {
	reqctx.send(metadata)
}
