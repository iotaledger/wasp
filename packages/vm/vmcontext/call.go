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
	ErrNotEnoughFees      = errors.New("not enough fees")
	ErrWrongRequestToken  = errors.New("wrong request token")
)

// CallContract
func (vmctx *VMContext) CallContract(contract coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error) {
	vmctx.log.Debugw("Call", "contract", contract, "epCode", epCode.String())

	rec, ok := vmctx.findContractByHname(contract)
	if !ok {
		return nil, ErrContractNotFound
	}
	proc, err := vmctx.processors.GetOrCreateProcessor(rec, vmctx.getBinary)
	if err != nil {
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

func (vmctx *VMContext) CallView(contractHname coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	vmctx.log.Debugw("CallView", "contract", contractHname, "epCode", epCode.String())

	rec, ok := vmctx.findContractByHname(contractHname)
	if !ok {
		return nil, fmt.Errorf("failed to find contract with index %d", contractHname)
	}

	proc, err := vmctx.processors.GetOrCreateProcessor(rec, vmctx.getBinary)
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
	vmctx.log.Debugf("mustCallFromRequest: %s -- %s\n", vmctx.reqRef.RequestID().String(), req.String())

	// handles request token and node fees
	remaining, err := vmctx.mustDefaultHandleTokens()
	if err != nil {
		// may be due to not enough rewards
		vmctx.log.Warnf("mustCallFromRequest: %v", err)
		return
	}

	// call contract from request context
	_, err = vmctx.CallContract(vmctx.reqHname, req.EntryPointCode(), req.Args(), remaining)

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
		// if no fees enabled, accrue the token to the caller
		fee := map[balance.Color]int64{
			balance.ColorIOTA: 1,
		}
		vmctx.creditToAccount(vmctx.reqRef.SenderAgentID(), accounts.NewColoredBalancesFromMap(fee))
		return transfer, nil
	}

	// handle fees
	if vmctx.contractRecord.NodeFee-1 > transfer.Balance(balance.ColorIOTA) {
		// fallback: not enough fees
		// accrue everything to the sender
		sender := vmctx.reqRef.SenderAgentID()
		vmctx.creditToAccount(sender, transfer)

		return accounts.NewColoredBalancesFromMap(nil), fmt.Errorf("not enough fees for request %s. Transfer accrued to %s",
			vmctx.reqRef.RequestID().Short(), sender.String())
	}
	// enough fees
	// accrue everything (including request token) to the chain owner
	fee := map[balance.Color]int64{
		balance.ColorIOTA: vmctx.contractRecord.NodeFee,
	}
	vmctx.creditToAccount(vmctx.ChainOwnerID(), accounts.NewColoredBalancesFromMap(fee))
	remaining := map[balance.Color]int64{
		balance.ColorIOTA: -vmctx.contractRecord.NodeFee + 1,
	}
	transfer.AddToMap(remaining)
	return accounts.NewColoredBalancesFromMap(remaining), nil
}

func (vmctx *VMContext) Params() codec.ImmutableCodec {
	return vmctx.getCallContext().params
}
