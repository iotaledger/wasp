package vmimpl

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

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

func (vmctx *vmContext) mustBeCalledFromContract(contract *coreutil.ContractInfo) {
	if vmctx.CurrentContractHname() != contract.Hname() {
		panic(fmt.Sprintf("%v: core contract '%s' expected", vm.ErrPrivilegedCallFailed, contract.Name))
	}
}

func (vmctx *vmContext) TryLoadContract(programHash hashing.HashValue) error {
	vmctx.mustBeCalledFromContract(root.Contract)
	vmtype, programBinary, err := execution.GetProgramBinary(vmctx, programHash)
	if err != nil {
		return err
	}
	return vmctx.task.Processors.NewProcessor(programHash, programBinary, vmtype)
}

func (vmctx *vmContext) CreateNewFoundry(scheme iotago.TokenScheme, metadata []byte) (uint32, uint64) {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	return vmctx.txbuilder.CreateNewFoundry(scheme, metadata)
}

func (vmctx *vmContext) DestroyFoundry(sn uint32) uint64 {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	return vmctx.txbuilder.DestroyFoundry(sn)
}

func (vmctx *vmContext) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	vmctx.mustBeCalledFromContract(accounts.Contract)
	out, _, _ := accounts.GetFoundryOutput(vmctx.State(), sn, vmctx.ChainID())
	nativeTokenID, err := out.NativeTokenID()
	if err != nil {
		panic(fmt.Errorf("internal: %w", err))
	}
	return vmctx.txbuilder.ModifyNativeTokenSupply(nativeTokenID, delta)
}

func (vmctx *vmContext) RetryUnprocessable(req isc.Request, blockIndex uint32, outputIndex uint16) {
	// set the "retry output ID" so that the correct output is used by the txbuilder
	oid := vmctx.getOutputID(blockIndex, outputIndex)
	retryReq := isc.NewRetryOnLedgerRequest(req.(isc.OnLedgerRequest), oid)
	vmctx.reqCtx.unprocessableToRetry = append(vmctx.reqCtx.unprocessableToRetry, retryReq)
}

func (vmctx *vmContext) SetBlockContext(bctx interface{}) {
	vmctx.blockContext[vmctx.CurrentContractHname()] = bctx
}

func (vmctx *vmContext) BlockContext() interface{} {
	return vmctx.blockContext[vmctx.CurrentContractHname()]
}

func (vmctx *vmContext) CallOnBehalfOf(caller isc.AgentID, target, entryPoint isc.Hname, params dict.Dict, allowance *isc.Assets) dict.Dict {
	vmctx.Debugf("CallOnBehalfOf: caller = %s, target = %s, entryPoint = %s, params = %s", caller.String(), target.String(), entryPoint.String(), params.String())
	return vmctx.callProgram(target, entryPoint, params, allowance, caller)
}

func (vmctx *vmContext) SetEVMFailed(tx *types.Transaction, receipt *types.Receipt) {
	vmctx.reqCtx.evmFailed = &evmFailed{
		tx:      tx,
		receipt: receipt,
	}
}
