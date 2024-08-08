package vmimpl

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func (reqctx *requestContext) mustBeCalledFromContract(contract *coreutil.ContractInfo) {
	if reqctx.CurrentContractHname() != contract.Hname() {
		panic(fmt.Sprintf("%v: core contract '%s' expected", vm.ErrPrivilegedCallFailed, contract.Name))
	}
}

func (reqctx *requestContext) TryLoadContract(programHash hashing.HashValue) error {
	reqctx.mustBeCalledFromContract(root.Contract)
	vmtype, programBinary, err := execution.GetProgramBinary(reqctx, programHash)
	if err != nil {
		return err
	}
	return reqctx.vm.task.Processors.NewProcessor(programHash, programBinary, vmtype)
}

func (reqctx *requestContext) CreateNewFoundry(scheme iotago.TokenScheme, metadata []byte) (uint32, uint64) {
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

func (reqctx *requestContext) MintNFT(addr *cryptolib.Address, immutableMetadata []byte, issuer *cryptolib.Address) (uint16, *iotago.NFTOutput) {
	reqctx.mustBeCalledFromContract(accounts.Contract)
	panic("refactor me: vmtxbuilder.MintNFT")
	// return reqctx.vm.txbuilder.MintNFT(addr, immutableMetadata, issuer)
}

func (reqctx *requestContext) RetryUnprocessable(req isc.Request, outputID sui.ObjectID) {
	retryReq := isc.NewRetryOnLedgerRequest(req.(isc.OnLedgerRequest), outputID)
	reqctx.unprocessableToRetry = append(reqctx.unprocessableToRetry, retryReq)
}

func (reqctx *requestContext) CallOnBehalfOf(caller isc.AgentID, msg isc.Message, allowance *isc.Assets) isc.CallArguments {
	reqctx.Debugf("CallOnBehalfOf: caller = %s, msg = %s", caller.String(), msg)
	return reqctx.callProgram(msg, allowance, caller)
}

func (reqctx *requestContext) SendOnBehalfOf(caller isc.ContractIdentity, metadata isc.RequestParameters) {
	reqctx.doSend(caller, metadata)
}
