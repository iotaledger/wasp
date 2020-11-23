package accountsc

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/cbalances"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize the init call
func initialize(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("accountsc.initialize.begin")
	state := ctx.State()
	if state.Get(VarStateInitialized) != nil {
		// can't be initialized twice
		return nil, fmt.Errorf("accountsc.initialize.fail: already_initialized")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	ctx.Eventf("accountsc.initialize.success hname = %s", Hname.String())
	return nil, nil
}

// getBalance returns colored balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func getBalance(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	ctx.Eventf("getBalance")
	aid, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamWrongOrNotFound
	}
	ctx.Eventf("getBalance for %s", aid.String())

	retMap, ok := GetAccountBalances(ctx.State(), *aid)
	ret := codec.NewCodec(dict.New())
	if !ok {
		return ret, nil
	}
	for col, bal := range retMap {
		ret.SetInt64(kv.Key(col[:]), bal)
	}
	return ret, nil
}

// getAccounts returns list of all accounts as keys of the ImmutableCodec
func getAccounts(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	return GetAccounts(ctx.State()), nil
}

// deposit moves transfer to the specified account
// Params:
// - ParamAgentID. default is ctx.Caller()
func deposit(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	state := ctx.State()
	MustCheckLedger(state, "accountsc.deposit.begin")
	defer MustCheckLedger(state, "accountsc.deposit.exit")

	ctx.Eventf("accountsc.deposit.begin")
	targetAgentID := ctx.Caller()
	aid, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if ok {
		targetAgentID = *aid
	}
	// funds currently are at the disposition of accountsc, they are moved to the target
	if !MoveBetweenAccounts(state, ctx.MyAgentID(), targetAgentID, ctx.Accounts().Incoming()) {
		return nil, fmt.Errorf("failed to deposit to %s", ctx.Caller().String())
	}
	ctx.Eventf("accountsc.deposit.success")
	return nil, nil
}

// moveOnChain moves funds on chain.
// Parameters:
// - ParamAgentID the target account
// - ParamColor color of the tokens. Default is iota color
// - ParamAmount the amount to move
func moveOnChain(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	state := ctx.State()
	MustCheckLedger(state, "accountsc.moveOnChain.begin")
	defer MustCheckLedger(state, "accountsc.moveOnChain.exit")

	ctx.Eventf("accountsc.moveOnChain.begin")
	params := ctx.Params()
	moveTo, ok, err := params.GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamWrongOrNotFound
	}
	amount, ok, err := params.GetInt64(ParamAmount)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamWrongOrNotFound
	}
	color, ok, err := params.GetColor(ParamColor)
	if err != nil {
		return nil, err
	}
	if !ok {
		*color = balance.ColorIOTA
	}
	move := cbalances.NewFromMap(map[balance.Color]int64{*color: amount})
	if !MoveBetweenAccounts(state, ctx.Caller(), *moveTo, move) {
		return nil, fmt.Errorf("failed to moveOnChain to %s: %s", moveTo.String(), move.String())
	}
	ctx.Eventf("accountsc.moveOnChain.success: %s", move.String())
	return nil, nil
}

// withdraw sends caller's funds to the caller
// different process for addresses and contracts as a caller
func withdraw(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	state := ctx.State()
	MustCheckLedger(state, "accountsc.withdraw.begin")
	defer MustCheckLedger(state, "accountsc.withdraw.exit")

	if state.Get(VarStateInitialized) == nil {
		return nil, fmt.Errorf("accountsc.initialize.fail: not_initialized")
	}
	caller := ctx.Caller()
	ctx.Eventf("accountsc.withdraw.begin: caller agentID: %s myContractId: %s", caller.String(), ctx.MyContractID().String())

	bals, ok := GetAccountBalances(state, caller)
	if !ok {
		return nil, fmt.Errorf("accountsc.withdraw.success. Inconsistency 1, empty account")
	}
	if caller.IsAddress() {
		// caller is address
		send := cbalances.NewFromMap(bals)
		addr := caller.MustAddress()
		if !DebitFromAccount(state, caller, send) {
			return nil, fmt.Errorf("accountsc.withdraw.success. Inconsistency 2, DebitFromAccount failed")
		}
		if !ctx.TransferToAddress(addr, send) {
			return nil, fmt.Errorf("accountsc.withdraw.fail: TransferToAddress failed")
		}
		// sent to address
		ctx.Eventf("accountsc.withdraw.success. Sent to address %s", addr.String())
		return nil, nil
	}
	// it is another contract. Deposit funds to another chain on caller's account
	// take it all tothe accountsc
	if !MoveBetweenAccounts(state, caller, ctx.MyAgentID(), cbalances.NewFromMap(bals)) {
		return nil, fmt.Errorf("accountsc.withdraw.success. Inconsistency 2, MoveBetweenAccounts failed")
	}
	// need one iota for request
	iotas, ok := bals[balance.ColorIOTA]
	if iotas <= 0 {
		return nil, fmt.Errorf("accountsc.withdraw.success. Inconsistency 3, empty iotas account")
	}
	bals[balance.ColorIOTA] = iotas - 1
	targetChain := caller.MustContractID().ChainID()
	par := dict.New()
	parCodec := codec.NewMustCodec(par)
	parCodec.SetAgentID(ParamAgentID, &caller)
	if !ctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: coretypes.NewContractID(targetChain, Hname),
		EntryPoint:       coretypes.Hn(FuncDeposit),
		Params:           par,
		Transfer:         cbalances.NewFromMap(bals),
	}) {
		return nil, fmt.Errorf("failed to post request")
	}
	ctx.Eventf("accountsc.withdraw.success")
	return nil, nil
}
