package vmcontext

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func (vmctx *VMContext) chainInfo() (coretypes.ChainID, coretypes.AgentID) {
	info, err := vmctx.Call(root.Interface.Hname(), coretypes.Hn(root.FuncGetInfo), nil, nil)
	if err != nil {
		panic(err)
	}
	chainID, ok, err := codec.DecodeChainID(info.MustGet(root.VarChainID))
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("inconsistency in the root 1")
	}
	if vmctx.chainID != chainID {
		panic("inconsistency in the root 2: vmctx.chainID != *chainID")
	}
	owner, ok, err := codec.DecodeAgentID(info.MustGet(root.VarChainOwnerID))
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("inconsistency in the root 3")
	}
	return chainID, owner
}

func (vmctx *VMContext) ChainID() coretypes.ChainID {
	return vmctx.chainID
}

func (vmctx *VMContext) ChainOwnerID() coretypes.AgentID {
	_, ret := vmctx.chainInfo()
	return ret
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

func (vmctx *VMContext) TransferToAddress(targetAddr address.Address, transfer coretypes.ColoredBalances) bool {
	privileged := vmctx.CurrentContractHname() == accountsc.Interface.Hname()
	fmt.Printf("TransferToAddress: %s privileged = %v\n", targetAddr.String(), privileged)
	if !privileged {
		// if caller is accoutsc, it must debit from account by itself
		if !accountsc.DebitFromAccount(vmctx.State(), vmctx.MyAgentID(), transfer) {
			return false
		}
	}
	return vmctx.txBuilder.TransferToAddress(targetAddr, transfer) == nil
}

// TransferCrossChain moves the whole transfer to another chain to the target account
// 1 request token should not be included into the transfer parameter but it is transferred automatically
// as a request token from the caller's account on top of specified transfer. It will be taken as a fee or accrued
// to the caller's account
// node fee is deducted from the transfer by the target
func (vmctx *VMContext) TransferCrossChain(targetAgentID coretypes.AgentID, targetChainID coretypes.ChainID, transfer coretypes.ColoredBalances) bool {
	if targetChainID == vmctx.ChainID() {
		return false
	}
	// the transfer is performed by the accountsc contract on another chain
	// it deposits received funds to the target on behalf of the caller
	par := dict.New()
	par.Set(accountsc.ParamAgentID, codec.EncodeAgentID(targetAgentID))
	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: coretypes.NewContractID(targetChainID, accountsc.Interface.Hname()),
		EntryPoint:       coretypes.Hn(accountsc.FuncDeposit),
		Params:           par,
		Transfer:         transfer,
	})
}

// PostRequest creates a request section in the transaction with specified parameters
// The transfer not include 1 iota for the request token but includes node fee, if eny
func (vmctx *VMContext) PostRequest(par vmtypes.NewRequestParams) bool {
	vmctx.log.Debugw("-- PostRequest",
		"target", par.TargetContractID.String(),
		"ep", par.EntryPoint.String(),
		"transfer", cbalances.Str(par.Transfer),
	)
	toAgentID := vmctx.MyAgentID()
	if !vmctx.debitFromAccount(toAgentID, cbalances.NewFromMap(map[balance.Color]int64{
		balance.ColorIOTA: 1,
	})) {
		vmctx.log.Debugf("-- PostRequest: not enough funds for request token")
		return false
	}
	if !vmctx.debitFromAccount(toAgentID, par.Transfer) {
		vmctx.log.Debugf("-- PostRequest: not enough funds")
		return false
	}
	reqSection := sctransaction.NewRequestSection(vmctx.CurrentContractHname(), par.TargetContractID, par.EntryPoint).
		WithTimelock(par.Timelock).
		WithTransfer(par.Transfer).
		WithArgs(par.Params)
	return vmctx.txBuilder.AddRequestSection(reqSection) == nil
}

func (vmctx *VMContext) PostRequestToSelf(reqCode coretypes.Hname, params dict.Dict) bool {
	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: vmctx.CurrentContractID(),
		EntryPoint:       reqCode,
		Params:           params,
	})
}

func (vmctx *VMContext) PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
	timelock := util.NanoSecToUnixSec(vmctx.timestamp) + delaySec

	return vmctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: vmctx.CurrentContractID(),
		EntryPoint:       entryPoint,
		Params:           args,
		Timelock:         timelock,
	})
}

func (vmctx *VMContext) EventPublisher() vm.ContractEventPublisher {
	return vm.NewContractEventPublisher(vmctx.CurrentContractID(), vmctx.log)
}

func (vmctx *VMContext) Request() *sctransaction.RequestRef {
	return &vmctx.reqRef
}
