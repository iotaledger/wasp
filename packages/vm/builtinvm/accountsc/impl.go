package accountsc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	accounts "github.com/iotaledger/wasp/packages/vm/balances"
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
func getBalance(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	ctx.Eventf("getBalance")
	aid, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamsAgentIDNotFound
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

func getAccounts(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	return GetAccounts(ctx.State()), nil
}

// deposit moves balances to the specified account, if any.
// if target account is not in parameters it is deposited to the caller's account
func deposit(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	targetAgentID := ctx.Caller()
	aid, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if ok {
		targetAgentID = *aid
	}
	// funds currently are at the disposition of accountsc, they are moved to the target
	if !MoveBetweenAccounts(ctx.State(), ctx.MyAgentID(), targetAgentID, ctx.Accounts().Incoming()) {
		return nil, fmt.Errorf("failed to deposit to %s", ctx.Caller().String())
	}
	return nil, nil
}

func moveOnChain(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	moveTo, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamsAgentIDNotFound
	}
	if !MoveBetweenAccounts(ctx.State(), ctx.MyAgentID(), *moveTo, ctx.Accounts().Incoming()) {
		return nil, fmt.Errorf("failed to moveOnChain to %s", moveTo.String())
	}
	return nil, nil
}

// withdraw sends caller's funds to the caller
// TODO with all kinds of callers
func withdraw(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	state := ctx.State()
	if state.Get(VarStateInitialized) == nil {
		return nil, fmt.Errorf("accountsc.initialize.fail: not_initialized")
	}
	caller := ctx.Caller()
	ctx.Eventf("accountsc.withdraw.begin: caller agentID: %s myContractId: %s", caller.String(), ctx.MyContractID().String())

	if !caller.IsAddress() {
		return nil, fmt.Errorf("accountsc.withdraw.fail: can't send tokens, must be an address. AgentID: %s", caller.String())
	}
	bals, ok := GetAccountBalances(state, caller)
	if !ok {
		return nil, fmt.Errorf("accountsc.withdraw.fail: account not found 1")
	}
	send := accounts.NewColoredBalancesFromMap(bals)
	if !DebitFromAccount(state, caller, send) {
		return nil, fmt.Errorf("accountsc.withdraw.fail: internal error 1")
	}
	if !ctx.TransferToAddress(caller.MustAddress(), send) {
		return nil, fmt.Errorf("accountsc.withdraw.fail: TransferToAddress failed")
	}
	ctx.Eventf("accountsc.withdraw.success")
	return nil, nil
}
