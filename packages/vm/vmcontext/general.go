package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm"
)

func (vmctx *VMContext) ChainID() *coretypes.ChainID {
	return &vmctx.chainID
}

func (vmctx *VMContext) ChainOwnerID() *coretypes.AgentID {
	return &vmctx.chainOwnerID
}

func (vmctx *VMContext) ContractCreator() *coretypes.AgentID {
	rec, ok := vmctx.findContractByHname(vmctx.CurrentContractHname())
	if !ok {
		vmctx.log.Panicf("can't find current contract")
	}
	return rec.Creator
}

func (vmctx *VMContext) CurrentContractHname() coretypes.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) MyAgentID() *coretypes.AgentID {
	return coretypes.NewAgentID(vmctx.ChainID().AsAddress(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) Minted() map[ledgerstate.Color]uint64 {
	if req, ok := vmctx.req.(*request.RequestOnLedger); ok {
		return req.MintedAmounts()
	}
	return nil
}

func (vmctx *VMContext) IsRequestContext() bool {
	return vmctx.getCallContext().isRequestContext
}

func (vmctx *VMContext) Caller() *coretypes.AgentID {
	ret := vmctx.getCallContext().caller
	return &ret
}

func (vmctx *VMContext) Timestamp() int64 {
	return vmctx.virtualState.Timestamp().UnixNano()
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

// PostRequest creates a request section in the transaction with specified parameters
// The transfer not include 1 iota for the request token but includes node fee, if eny
//func (vmctx *VMContext) PostRequest(par coretypes.PostRequestParams) bool {
//	vmctx.log.Debugw("-- PostRequestSync",
//		"target", par.TargetContractID.String(),
//		"ep", par.VMProcessorEntryPoint.String(),
//		"transfer", par.Transfer.String(),
//	)
//	myAgentID := vmctx.MyAgentID()
//	if !vmctx.debitFromAccount(myAgentID, par.Transfer) {
//		vmctx.log.Debugf("-- PostRequestSync: not enough funds")
//		return false
//	}
//	reqParams := requestargs.New(nil)
//	reqParams.AddEncodeSimpleMany(par.Params)
//	reqSection := sctransaction_old.NewRequestSection(vmctx.CurrentContractHname(), par.TargetContractID, par.VMProcessorEntryPoint).
//		WithTimeLock(par.TimeLock).
//		WithTransfer(par.Transfer).
//		WithArgs(reqParams)
//	return vmctx.txBuilder.AddRequestSection(reqSection) == nil
//}
//
//func (vmctx *VMContext) PostRequestToSelf(reqCode coretypes.Hname, params dict.Dict) bool {
//	return vmctx.PostRequest(coretypes.PostRequestParams{
//		TargetContractID: vmctx.CurrentContractID(),
//		VMProcessorEntryPoint:       reqCode,
//		Params:           params,
//	})
//}
//
//func (vmctx *VMContext) PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
//	timelock := util.NanoSecToUnixSec(vmctx.timestamp) + delaySec
//
//	return vmctx.PostRequest(coretypes.PostRequestParams{
//		TargetContractID: vmctx.CurrentContractID(),
//		VMProcessorEntryPoint:       entryPoint,
//		Params:           args,
//		TimeLock:         timelock,
//	})
//}

func (vmctx *VMContext) EventPublisher() vm.ContractEventPublisher {
	return vm.NewContractEventPublisher(vmctx.ChainID(), vmctx.CurrentContractHname(), vmctx.log)
}

func (vmctx *VMContext) RequestID() coretypes.RequestID {
	return vmctx.req.ID()
}

const maxParamSize = 512

func (vmctx *VMContext) Send(target ledgerstate.Address, tokens *ledgerstate.ColoredBalances, metadata *coretypes.SendMetadata, options ...coretypes.SendOptions) bool {
	if tokens == nil || tokens.Size() == 0 {
		vmctx.log.Errorf("Send: transfer can't be empty")
		return false
	}
	data := request.NewRequestMetadata().
		WithSender(vmctx.CurrentContractHname())
	if metadata != nil {
		var args requestargs.RequestArgs
		if metadata.Args != nil && len(metadata.Args) > 0 {
			var opt map[kv.Key][]byte
			args, opt = requestargs.NewOptimizedRequestArgs(metadata.Args, maxParamSize)
			if len(opt) > 0 {
				// some parameters  too big
				vmctx.log.Errorf("Send: too big data in parameters")
				return false
			}
		}
		data.WithTarget(metadata.TargetContract).
			WithEntryPoint(metadata.EntryPoint).
			WithArgs(args)
	}
	sourceAccount := vmctx.adjustAccount(vmctx.MyAgentID())
	if !vmctx.debitFromAccount(sourceAccount, tokens) {
		return false
	}
	err := vmctx.txBuilder.AddExtendedOutputSpend(target, data.Bytes(), tokens.Map())
	if err != nil {
		vmctx.log.Errorf("Send: %v", err)
		return false
	}
	return true
}
