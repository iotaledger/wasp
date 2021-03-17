package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm"
)

func (vmctx *VMContext) ChainID() coretypes.ChainID {
	return vmctx.chainID
}

func (vmctx *VMContext) ChainOwnerID() *coretypes.AgentID {
	return &vmctx.chainOwnerID
}

func (vmctx *VMContext) ContractCreator() *coretypes.AgentID {
	rec, ok := vmctx.findContractByHname(vmctx.CurrentContractHname())
	if !ok {
		vmctx.log.Panicf("can't find current contract")
	}
	ret := rec.Creator
	return &ret
}

func (vmctx *VMContext) CurrentContractHname() coretypes.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) CurrentContractID() *coretypes.ContractID {
	ret := coretypes.NewContractID(vmctx.ChainID(), vmctx.CurrentContractHname())
	return ret
}

func (vmctx *VMContext) MyAgentID() *coretypes.AgentID {
	return coretypes.NewAgentIDFromContractID(vmctx.CurrentContractID())
}

func (vmctx *VMContext) IsRequestContext() bool {
	return vmctx.getCallContext().isRequestContext
}

func (vmctx *VMContext) Caller() *coretypes.AgentID {
	ret := vmctx.getCallContext().caller
	return &ret
}

func (vmctx *VMContext) Timestamp() int64 {
	return vmctx.timestamp
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

// PostRequest creates a request section in the transaction with specified parameters
// The transfer not include 1 iota for the request token but includes node fee, if eny
//func (vmctx *VMContext) PostRequest(par coretypes.PostRequestParams) bool {
//	vmctx.log.Debugw("-- PostRequestSync",
//		"target", par.TargetContractID.String(),
//		"ep", par.EntryPoint.String(),
//		"transfer", par.Transfer.String(),
//	)
//	myAgentID := vmctx.MyAgentID()
//	if !vmctx.debitFromAccount(myAgentID, par.Transfer) {
//		vmctx.log.Debugf("-- PostRequestSync: not enough funds")
//		return false
//	}
//	reqParams := requestargs.New(nil)
//	reqParams.AddEncodeSimpleMany(par.Params)
//	reqSection := sctransaction_old.NewRequestSection(vmctx.CurrentContractHname(), par.TargetContractID, par.EntryPoint).
//		WithTimeLock(par.TimeLock).
//		WithTransfer(par.Transfer).
//		WithArgs(reqParams)
//	return vmctx.txBuilder.AddRequestSection(reqSection) == nil
//}
//
//func (vmctx *VMContext) PostRequestToSelf(reqCode coretypes.Hname, params dict.Dict) bool {
//	return vmctx.PostRequest(coretypes.PostRequestParams{
//		TargetContractID: vmctx.CurrentContractID(),
//		EntryPoint:       reqCode,
//		Params:           params,
//	})
//}
//
//func (vmctx *VMContext) PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
//	timelock := util.NanoSecToUnixSec(vmctx.timestamp) + delaySec
//
//	return vmctx.PostRequest(coretypes.PostRequestParams{
//		TargetContractID: vmctx.CurrentContractID(),
//		EntryPoint:       entryPoint,
//		Params:           args,
//		TimeLock:         timelock,
//	})
//}

func (vmctx *VMContext) EventPublisher() vm.ContractEventPublisher {
	return vm.NewContractEventPublisher(*vmctx.CurrentContractID(), vmctx.log)
}

func (vmctx *VMContext) RequestID() ledgerstate.OutputID {
	return vmctx.req.Output().ID()
}
