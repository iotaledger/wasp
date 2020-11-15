package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	accounts "github.com/iotaledger/wasp/packages/vm/balances"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
)

// CallContract
func (vmctx *VMContext) CallContract(contract coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error) {
	vmctx.log.Debugw("Call", "contract", contract, "epCode", epCode.String())

	rec, ok := vmctx.findContractByHname(contract)
	if !ok {
		return nil, fmt.Errorf("failed to find contract with hname %s", contract.String())
	}

	proc, err := vmctx.getProcessor(rec)
	if err != nil {
		return nil, err
	}

	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, fmt.Errorf("can't find entry point for entry point '%s'", epCode.String())
	}

	if err := vmctx.PushCallContext(contract, params, transfer); err != nil {
		return nil, err
	}
	defer vmctx.PopCallContext()

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

	if err := vmctx.PushCallContext(contractHname, params, nil); err != nil {
		return nil, err
	}
	defer vmctx.PopCallContext()

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
		transfer = vmctx.mustMoveFees()
	}
	_, err := vmctx.CallContract(vmctx.reqHname, req.EntryPointCode(), req.Args(), transfer)
	if err != nil {
		vmctx.log.Warnf("mustCallFromRequest: %v", err)
	}
}

// mustFallbackNotEnoughFees calls fallback reaction in case fees iotas not enough for fees
func (vmctx *VMContext) mustFallbackNotEnoughFees() {
	transfer := vmctx.reqRef.RequestSection().Transfer()
	if _, err := vmctx.CallContract(
		accountsc.Hname,
		accountsc.EntryPointFuncFallbackShortOfFees,
		vmctx.reqRef.RequestSection().Args(),
		transfer); err != nil {
		vmctx.log.Panicf("executing fallback due to not enough fees': %v", err)
		return
	}
	vmctx.log.Warnf("not enough fees for request %s", vmctx.reqRef.RequestID().Short())
}

// mustMoveFees call accountsc contract to move fees to destination
func (vmctx *VMContext) mustMoveFees() coretypes.ColoredBalances {
	par := codec.NewCodec(dict.NewDict())
	par.SetAgentID(accountsc.ParamAgentID, &vmctx.accrueFeesTo)
	fees := accounts.NewColoredBalancesFromMap(map[balance.Color]int64{
		balance.ColorIOTA: vmctx.contractRecord.NodeFee,
	})
	if _, err := vmctx.CallContract(accountsc.Hname, accountsc.EntryPointMoveTokens, par, fees); err != nil {
		vmctx.log.Panicf("moving fees: %v", err)
	}
	transfer := vmctx.reqRef.RequestSection().Transfer()
	debit := map[balance.Color]int64{
		balance.ColorIOTA: -vmctx.contractRecord.NodeFee,
	}
	transfer.AddToMap(debit)
	return accounts.NewColoredBalancesFromMap(debit)
}

func (vmctx *VMContext) Params() codec.ImmutableCodec {
	return vmctx.getCallContext().params
}
