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

func (vmctx *VMContext) CurrentContractHname() coretypes.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) IsRequestContext() bool {
	return vmctx.getCallContext().isRequestContext
}

func (vmctx *VMContext) CurrentCaller() coretypes.AgentID {
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
	if vmctx.getCallContext().contract == accountsc.Hname {
		reqSection := sctransaction.NewRequestSection(par.SenderContractHname, par.TargetContractID, par.EntryPoint).
			WithTimelock(par.Timelock).
			WithTransfer(par.Transfer).
			WithArgs(par.Params)
		vmctx.txBuilder.AddRequestSection(reqSection)
		return true
	}
	par1 := codec.NewCodec(par.Params.Clone())
	par1.SetHname(accountsc.ParamSenderContractHname__, vmctx.CurrentContractHname())
	par1.SetContractID(accountsc.ParamTargetContractID__, &par.TargetContractID)
	par1.SetHname(accountsc.ParamTargetEntryPoint__, par.EntryPoint)
	_, err := vmctx.CallContract(accountsc.Hname, accountsc.EntryPointPostRequest, par1, par.Transfer)
	return err == nil
}

func (vmctx *VMContext) SendRequestToSelf(reqCode coretypes.Hname, params dict.Dict) bool {
	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: coretypes.NewContractID(vmctx.chainID, vmctx.CurrentContractHname()),
		EntryPoint:       reqCode,
		Params:           params,
	})
}

func (vmctx *VMContext) SendRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
	timelock := util.NanoSecToUnixSec(vmctx.timestamp) + delaySec

	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: coretypes.NewContractID(vmctx.chainID, vmctx.CurrentContractHname()),
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
