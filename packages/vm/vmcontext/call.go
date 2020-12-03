package vmcontext

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

var (
	ErrContractNotFound   = errors.New("contract not found")
	ErrEntryPointNotFound = errors.New("entry point not found")
	ErrProcessorNotFound  = errors.New("VM not found. Internal error")
	ErrNotEnoughFees      = errors.New("not enough fees")
	ErrWrongRequestToken  = errors.New("wrong request token")
)

// Call
func (vmctx *VMContext) Call(contract coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error) {
	vmctx.log.Debugw("Call", "contract", contract, "epCode", epCode.String())

	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == coretypes.EntryPointInit &&
		contract != root.Interface.Hname() &&
		vmctx.CurrentContractHname() != root.Interface.Hname() {
		return nil, fmt.Errorf("attempt to call init not from root contract")
	}

	rec, ok := vmctx.findContractByHname(contract)
	if !ok {
		return nil, ErrContractNotFound
	}
	proc, err := vmctx.processors.GetOrCreateProcessor(rec, vmctx.getBinary)
	if err != nil {
		return nil, err
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, ErrEntryPointNotFound
	}

	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		// passing nil as transfer: calling to view should not have effect on chain ledger
		if err := vmctx.pushCallContextWithTransfer(contract, params, nil); err != nil {
			return nil, err
		}
		defer vmctx.popCallContext()

		return ep.CallView(NewSandboxView(vmctx))
	}
	if err := vmctx.pushCallContextWithTransfer(contract, params, transfer); err != nil {
		return nil, err
	}
	defer vmctx.popCallContext()
	return ep.Call(NewSandbox(vmctx))
}

// mustCallFromRequest is called for each request from the VM loop
func (vmctx *VMContext) mustCallFromRequest() {
	req := vmctx.reqRef.RequestSection()
	vmctx.log.Debugf("mustCallFromRequest: %s -- %s\n", vmctx.reqRef.RequestID().String(), req.String())

	// handles request token and node fees
	remaining, err := vmctx.mustDefaultHandleTokens()
	if err != nil {
		// may be due to not enough rewards
		vmctx.log.Warnf("mustCallFromRequest: %v", err)
		return
	}

	// call contract from request context
	_, err = vmctx.Call(vmctx.reqHname, req.EntryPointCode(), req.Args(), remaining)

	switch err {
	case nil:
		return
	case ErrContractNotFound, ErrEntryPointNotFound, ErrProcessorNotFound:
		// if sent to the wrong contract or entry point, accrue the transfer to the sender' account on the chain
		// the sender can withdraw it at any time
		// TODO more sophisticated policy
		vmctx.creditToAccount(vmctx.reqRef.SenderAgentID(), remaining)
	default:
		vmctx.log.Errorf("mustCallFromRequest: %v reqid: %s", err, vmctx.reqRef.RequestID().String())
	}
}

// mustDefaultHandleTokens:
// - handles request token
// - handles node fee, including fallback if not enough
func (vmctx *VMContext) mustDefaultHandleTokens() (coretypes.ColoredBalances, error) {
	transfer := vmctx.reqRef.RequestSection().Transfer()
	reqColor := balance.Color(vmctx.reqRef.Tx.ID())

	// handle request token
	if vmctx.txBuilder.Balance(reqColor) == 0 {
		// must be checked before, while validating transaction
		vmctx.log.Panicf("request token not found: %s", reqColor.String())
	}
	if !vmctx.txBuilder.Erase1TokenToChain(reqColor) {
		vmctx.log.Panicf("internal error: can't destroy request token not found: %s", reqColor.String())
	}
	if vmctx.contractRecord.NodeFee == 0 {
		fmt.Printf("fees disabled, credit 1 iota to %s\n", vmctx.reqRef.SenderAgentID())
		// if no fees enabled, accrue the token to the caller
		fee := map[balance.Color]int64{
			balance.ColorIOTA: 1,
		}
		vmctx.creditToAccount(vmctx.reqRef.SenderAgentID(), cbalances.NewFromMap(fee))
		return transfer, nil
	}

	// handle fees
	if vmctx.contractRecord.NodeFee-1 > transfer.Balance(balance.ColorIOTA) {
		// fallback: not enough fees
		// accrue everything to the sender
		sender := vmctx.reqRef.SenderAgentID()
		vmctx.creditToAccount(sender, transfer)

		return cbalances.NewFromMap(nil), fmt.Errorf("not enough fees for request %s. Transfer accrued to %s",
			vmctx.reqRef.RequestID().Short(), sender.String())
	}
	// enough fees
	// accrue everything (including request token) to the chain owner
	fee := map[balance.Color]int64{
		balance.ColorIOTA: vmctx.contractRecord.NodeFee,
	}
	vmctx.creditToAccount(vmctx.ChainOwnerID(), cbalances.NewFromMap(fee))
	remaining := map[balance.Color]int64{
		balance.ColorIOTA: -vmctx.contractRecord.NodeFee + 1,
	}
	transfer.AddToMap(remaining)
	return cbalances.NewFromMap(remaining), nil
}

func (vmctx *VMContext) Params() codec.ImmutableCodec {
	return vmctx.getCallContext().params
}
