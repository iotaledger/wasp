package vmcontext

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
)

func (vmctx *VMContext) ChainID() coretypes.ChainID {
	return vmctx.chainID
}

func (vmctx *VMContext) ChainOwnerID() coretypes.AgentID {
	return vmctx.chainOwnerID
}

func (vmctx *VMContext) ContractCreator() coretypes.AgentID {
	rec, ok := vmctx.findContractByHname(vmctx.CurrentContractHname())
	if !ok {
		vmctx.log.Panicf("can't find current contract")
	}
	return rec.Creator
}

func (vmctx *VMContext) CurrentContractHname() coretypes.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) CurrentContractID() coretypes.ContractID {
	return coretypes.NewContractID(vmctx.ChainID(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) MyAgentID() coretypes.AgentID {
	return coretypes.NewAgentIDFromContractID(vmctx.CurrentContractID())
}

func (vmctx *VMContext) IsRequestContext() bool {
	return vmctx.getCallContext().isRequestContext
}

func (vmctx *VMContext) Caller() coretypes.AgentID {
	return vmctx.getCallContext().caller
}

func (vmctx *VMContext) Timestamp() int64 {
	return vmctx.timestamp
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

// PostRequest creates a request section in the transaction with specified parameters
// The transfer not include 1 iota for the request token but includes node fee, if eny
func (vmctx *VMContext) PostRequest(par coretypes.PostRequestParams) bool {
	vmctx.log.Debugw("-- PostRequestSync",
		"target", par.TargetContractID.String(),
		"ep", par.EntryPoint.String(),
		"transfer", cbalances.Str(par.Transfer),
	)
	myAgentID := vmctx.MyAgentID()
	if !vmctx.debitFromAccount(myAgentID, cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 1,
	})) {
		vmctx.log.Debugf("-- PostRequestSync: not enough funds for request token")
		return false
	}
	if !vmctx.debitFromAccount(myAgentID, par.Transfer) {
		vmctx.log.Debugf("-- PostRequestSync: not enough funds")
		return false
	}
	reqParams := requestargs.New(nil)
	reqParams.AddEncodeSimpleMany(par.Params)
	reqSection := sctransaction.NewRequestSection(vmctx.CurrentContractHname(), par.TargetContractID, par.EntryPoint).
		WithTimelock(par.TimeLock).
		WithTransfer(par.Transfer).
		WithArgs(reqParams)
	return vmctx.txBuilder.AddRequestSection(reqSection) == nil
}

func (vmctx *VMContext) PostRequestToSelf(reqCode coretypes.Hname, params dict.Dict) bool {
	return vmctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: vmctx.CurrentContractID(),
		EntryPoint:       reqCode,
		Params:           params,
	})
}

func (vmctx *VMContext) PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
	timelock := util.NanoSecToUnixSec(vmctx.timestamp) + delaySec

	return vmctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: vmctx.CurrentContractID(),
		EntryPoint:       entryPoint,
		Params:           args,
		TimeLock:         timelock,
	})
}

func (vmctx *VMContext) EventPublisher() vm.ContractEventPublisher {
	return vm.NewContractEventPublisher(vmctx.CurrentContractID(), vmctx.log)
}

func (vmctx *VMContext) RequestID() coretypes.RequestID {
	return *vmctx.reqRef.RequestID()
}

func (vmctx *VMContext) NumFreeMinted() int64 {
	return vmctx.reqRef.Tx.MustProperties().NumFreeMintedTokens()
}
