package vmcontext

import (
	"fmt"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/plugins/publisher"
)

func (vmctx *VMContext) Accounts() vmtypes.Accounts {
	return vmctx
}

func (vmctx *VMContext) ChainID() coretypes.ChainID {
	return vmctx.chainID
}

func (vmctx *VMContext) ChainOwnerID() coretypes.AgentID {
	return coretypes.NewAgentIDFromContractID(coretypes.NewContractID(vmctx.ChainID(), accountsc.Hname))
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

func (vmctx *VMContext) Log() *logger.Logger {
	return vmctx.log
}

func (vmctx *VMContext) PostRequest(par vmtypes.NewRequestParams) bool {
	if !accountsc.DebitFromAccount(codec.NewMustCodec(vmctx), vmctx.MyAgentID(), par.Transfer) {
		return false
	}
	reqSection := sctransaction.NewRequestSection(vmctx.CurrentContractHname(), par.TargetContractID, par.EntryPoint).
		WithTimelock(par.Timelock).
		WithTransfer(par.Transfer).
		WithArgs(par.Params)
	return vmctx.txBuilder.AddRequestSection(reqSection) == nil
}

func (vmctx *VMContext) SendRequestToSelf(reqCode coretypes.Hname, params dict.Dict) bool {
	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: vmctx.CurrentContractID(),
		EntryPoint:       reqCode,
		Params:           params,
	})
}

func (vmctx *VMContext) SendRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
	timelock := util.NanoSecToUnixSec(vmctx.timestamp) + delaySec

	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: vmctx.CurrentContractID(),
		EntryPoint:       entryPoint,
		Params:           args,
		Timelock:         timelock,
	})
}

func (vmctx *VMContext) Publish(msg string) {
	publisher.Publish("vmmsg", vmctx.chainID.String(), fmt.Sprintf("%d", vmctx.CurrentContractHname()), msg)
}

func (vmctx *VMContext) Publishf(format string, args ...interface{}) {
	publisher.Publish("vmmsg", vmctx.chainID.String(), fmt.Sprintf("%d", vmctx.CurrentContractHname()), fmt.Sprintf(format, args...))
}

func (vmctx *VMContext) Request() *sctransaction.RequestRef {
	return &vmctx.reqRef
}
