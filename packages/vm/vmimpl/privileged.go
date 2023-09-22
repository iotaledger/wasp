package vmimpl

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
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
	return reqctx.vm.txbuilder.CreateNewFoundry(scheme, metadata)
}

func (reqctx *requestContext) DestroyFoundry(sn uint32) uint64 {
	reqctx.mustBeCalledFromContract(accounts.Contract)
	return reqctx.vm.txbuilder.DestroyFoundry(sn)
}

func (reqctx *requestContext) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	reqctx.mustBeCalledFromContract(accounts.Contract)
	out, _ := accounts.GetFoundryOutput(reqctx.contractStateWithGasBurn(), sn, reqctx.ChainID())
	nativeTokenID, err := out.NativeTokenID()
	if err != nil {
		panic(fmt.Errorf("internal: %w", err))
	}
	return reqctx.vm.txbuilder.ModifyNativeTokenSupply(nativeTokenID, delta)
}

func (reqctx *requestContext) MintNFT(addr iotago.Address, immutableMetadata []byte, issuer iotago.Address) (uint16, *iotago.NFTOutput) {
	reqctx.mustBeCalledFromContract(accounts.Contract)
	return reqctx.vm.txbuilder.MintNFT(addr, immutableMetadata, issuer)
}

func (reqctx *requestContext) RetryUnprocessable(req isc.Request, outputID iotago.OutputID) {
	retryReq := isc.NewRetryOnLedgerRequest(req.(isc.OnLedgerRequest), outputID)
	reqctx.unprocessableToRetry = append(reqctx.unprocessableToRetry, retryReq)
}

func (reqctx *requestContext) CallOnBehalfOf(caller isc.AgentID, target, entryPoint isc.Hname, params dict.Dict, allowance *isc.Assets) dict.Dict {
	reqctx.Debugf("CallOnBehalfOf: caller = %s, target = %s, entryPoint = %s, params = %s", caller.String(), target.String(), entryPoint.String(), params.String())
	return reqctx.callProgram(target, entryPoint, params, allowance, caller)
}

func (reqctx *requestContext) SendOnBehalfOf(caller isc.ContractIdentity, metadata isc.RequestParameters) {
	reqctx.doSend(caller, metadata)
}
