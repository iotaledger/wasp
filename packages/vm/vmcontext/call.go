package vmcontext

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	accounts "github.com/iotaledger/wasp/packages/vm/balances"
)

var (
	ErrContractNotFound   = errors.New("contract not found")
	ErrEntryPointNotFound = errors.New("entry point not found")
	ErrProcessorNotFound  = errors.New("VM not found. Internal error")
)

// CallContract
func (vmctx *VMContext) CallContract(contract coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error) {
	vmctx.log.Debugw("Call", "contract", contract, "epCode", epCode.String())

	rec, ok := vmctx.findContractByHname(contract)
	if !ok {
		return nil, ErrContractNotFound
	}
	proc, err := vmctx.getProcessor(rec)
	if err != nil {
		vmctx.log.Errorf("CallContract.getProcessor: %v", err)
		return nil, ErrProcessorNotFound
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, ErrEntryPointNotFound
	}

	if err := vmctx.pushCallContextWithTransfer(contract, params, transfer); err != nil {
		return nil, err
	}
	defer vmctx.popCallContext()

	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		return ep.CallView(NewSandboxView(vmctx))
	}
	return ep.Call(NewSandbox(vmctx))
}

// CallContract
func (vmctx *VMContext) CallView(contractHname coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	vmctx.log.Debugw("CallView", "contract", contractHname, "epCode", epCode.String())

	rec, ok := vmctx.findContractByHname(contractHname)
	if !ok {
		return nil, fmt.Errorf("failed to find contract with index %d", contractHname)
	}

	proc, err := vmctx.getProcessor(rec)
	if err != nil {
		return nil, err
	}

	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, fmt.Errorf("can't find entry point for entry point '%s'", epCode.String())
	}

	if err := vmctx.pushCallContextWithTransfer(contractHname, params, nil); err != nil {
		return nil, err
	}
	defer vmctx.popCallContext()

	if !ep.IsView() {
		return nil, fmt.Errorf("only view entry point can be called in this context")
	}
	return ep.CallView(NewSandboxView(vmctx))
}

// mustCallFromRequest is called for each request from the VM loop
func (vmctx *VMContext) mustCallFromRequest() {
	req := vmctx.reqRef.RequestSection()
	transfer := req.Transfer()
	vmctx.log.Debugf("mustCallFromRequest: %s -- %s\n", vmctx.reqRef.RequestID().String(), req.String())

	vmctx.txBuilder.RequestProcessed(*vmctx.reqRef.RequestID())
	if vmctx.contractRecord.NodeFee > 0 {
		// handle node fees
		if transfer.Balance(balance.ColorIOTA) < vmctx.contractRecord.NodeFee {
			vmctx.mustFallbackNotEnoughFees()
			return
		}
		transfer = vmctx.mustCreditFees()
	}
	_, err := vmctx.CallContract(vmctx.reqHname, req.EntryPointCode(), req.Args(), transfer)
	switch err {
	case nil:
		return
	case ErrContractNotFound, ErrEntryPointNotFound, ErrProcessorNotFound:
		// if sent to the wrong contract or entry point, accrue the transfer to the sender' account on the chain
		// the sender can withdraw it at any time
		// TODO more sophisticated policy
		sender := vmctx.reqRef.SenderAgentID()
		vmctx.creditToAccount(sender, transfer)
	default:
		vmctx.log.Warnf("mustCallFromRequest: %v", err)
	}
}

// mustFallbackNotEnoughFees calls fallback reaction in case fees iotas not enough for fees
func (vmctx *VMContext) mustFallbackNotEnoughFees() {
	transfer := vmctx.reqRef.RequestSection().Transfer()
	// move all tokens to the caller's account
	// TODO more sophisticated policy
	sender := vmctx.reqRef.SenderAgentID()
	vmctx.creditToAccount(sender, transfer)
	vmctx.log.Warnf("not enough fees for request %s", vmctx.reqRef.RequestID().Short())
}

// mustCreditFees adds fees to chain owner's account.
// Returns remaining transfer
func (vmctx *VMContext) mustCreditFees() coretypes.ColoredBalances {
	transfer := vmctx.reqRef.RequestSection().Transfer()
	if vmctx.contractRecord.NodeFee == 0 || transfer.Balance(balance.ColorIOTA) < vmctx.contractRecord.NodeFee {
		vmctx.log.Panicf("mustCreditFees: should not be called")
	}
	fee := map[balance.Color]int64{
		balance.ColorIOTA: vmctx.contractRecord.NodeFee,
	}
	vmctx.creditToAccount(vmctx.ChainOwnerID(), accounts.NewColoredBalancesFromMap(fee))
	remaining := map[balance.Color]int64{
		balance.ColorIOTA: -vmctx.contractRecord.NodeFee,
	}
	transfer.AddToMap(remaining)
	return accounts.NewColoredBalancesFromMap(remaining)
}

func (vmctx *VMContext) Params() codec.ImmutableCodec {
	return vmctx.getCallContext().params
}
