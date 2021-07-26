package vmcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm"
)

func (vmctx *VMContext) ChainID() *iscp.ChainID {
	return &vmctx.chainID
}

func (vmctx *VMContext) ChainOwnerID() *iscp.AgentID {
	return &vmctx.chainOwnerID
}

func (vmctx *VMContext) ContractCreator() *iscp.AgentID {
	rec, ok := vmctx.findContractByHname(vmctx.CurrentContractHname())
	if !ok {
		vmctx.log.Panicf("can't find current contract")
	}
	return rec.Creator
}

func (vmctx *VMContext) CurrentContractHname() iscp.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) MyAgentID() *iscp.AgentID {
	return iscp.NewAgentID(vmctx.ChainID().AsAddress(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) Minted() colored.Balances {
	if req, ok := vmctx.req.(*request.OnLedger); ok {
		return req.MintedAmounts()
	}
	return nil
}

func (vmctx *VMContext) IsRequestContext() bool {
	return vmctx.getCallContext().isRequestContext
}

func (vmctx *VMContext) Caller() *iscp.AgentID {
	ret := vmctx.getCallContext().caller
	return &ret
}

func (vmctx *VMContext) Timestamp() int64 {
	return vmctx.virtualState.Timestamp().UnixNano()
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

// Post1Request creates a request section in the transaction with specified parameters
// The transfer not include 1 iota for the request token but includes node fee, if eny
//func (vmctx *VMContext) Post1Request(par iscp.PostRequestParams) bool {
//	vmctx.log.Debugw("-- PostRequestSync",
//		"target", par.TargetContractID.String(),
//		"ep", par.VMProcessorEntryPoint.String(),
//		"transfer", par.Tokens.String(),
//	)
//	myAgentID := vmctx.MyAgentID()
//	if !vmctx.debitFromAccount(myAgentID, par.Tokens) {
//		vmctx.log.Debugf("-- PostRequestSync: not enough funds")
//		return false
//	}
//	reqParams := requestargs.New(nil)
//	reqParams.AddEncodeSimpleMany(par.Params)
//	reqSection := sctransaction_old.NewRequestSection(vmctx.CurrentContractHname(), par.TargetContractID, par.VMProcessorEntryPoint).
//		WithTimeLock(par.TimeLock).
//		WithTransfer(par.Tokens).
//		WithArgs(reqParams)
//	return vmctx.txBuilder.AddRequestSection(reqSection) == nil
//}
//
//func (vmctx *VMContext) PostRequestToSelf(reqCode iscp.Hname, params dict.Dict) bool {
//	return vmctx.Post1Request(iscp.PostRequestParams{
//		TargetContractID: vmctx.CurrentContractID(),
//		VMProcessorEntryPoint:       reqCode,
//		Params:           params,
//	})
//}
//
//func (vmctx *VMContext) PostRequestToSelfWithDelay(entryPoint iscp.Hname, args dict.Dict, delaySec uint32) bool {
//	timelock := util.NanoSecToUnixSec(vmctx.timestamp) + delaySec
//
//	return vmctx.Post1Request(iscp.PostRequestParams{
//		TargetContractID: vmctx.CurrentContractID(),
//		VMProcessorEntryPoint:       entryPoint,
//		Params:           args,
//		TimeLock:         timelock,
//	})
//}

func (vmctx *VMContext) EventPublisher() vm.ContractEventPublisher {
	return vm.NewContractEventPublisher(vmctx.ChainID(), vmctx.CurrentContractHname(), vmctx.log)
}

func (vmctx *VMContext) RequestID() iscp.RequestID {
	return vmctx.req.ID()
}

const maxParamSize = 512

// TODO implement send options
//goland:noinspection GoUnusedParameter
func (vmctx *VMContext) Send(target ledgerstate.Address, tokens colored.Balances, metadata *iscp.SendMetadata, options ...iscp.SendOptions) bool {
	if tokens == nil || len(tokens) == 0 {
		vmctx.log.Errorf("Send: transfer can't be empty")
		return false
	}
	data := request.NewMetadata().
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
	err := vmctx.txBuilder.AddExtendedOutputSpend(target, data.Bytes(), colored.ToL1Map(tokens))
	if err != nil {
		vmctx.log.Errorf("Send: %v", err)
		return false
	}
	return true
}

// - anchor properties
func (vmctx *VMContext) StateAddress() ledgerstate.Address {
	return vmctx.chainInput.GetStateAddress()
}

func (vmctx *VMContext) GoverningAddress() ledgerstate.Address {
	return vmctx.chainInput.GetGoverningAddress()
}

func (vmctx *VMContext) StateIndex() uint32 {
	return vmctx.chainInput.GetStateIndex()
}

func (vmctx *VMContext) StateHash() hashing.HashValue {
	var h hashing.HashValue
	h, _ = hashing.HashValueFromBytes(vmctx.chainInput.GetStateData())
	return h
}

func (vmctx *VMContext) OutputID() ledgerstate.OutputID {
	return vmctx.chainInput.ID()
}
