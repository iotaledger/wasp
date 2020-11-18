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
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		// can't be initialized twice
		return nil, fmt.Errorf("accountsc.initialize.fail: already_initialized")
	}
	ctx.Eventf("accountsc.initialize.success hname = %s", Hname.String())
	return nil, nil
}

// getBalance returns colored balances of the account belonging to the AgentID
func getBalance(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	aid, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamsAgentIDNotFound
	}
	retMap, ok := GetAccountBalances(ctx.State(), *aid)
	if !ok {
		return nil, fmt.Errorf("getBalance: fail")
	}
	ret := codec.NewCodec(dict.New())
	for col, bal := range retMap {
		ret.SetInt64(kv.Key(col[:]), bal)
	}
	return ret, nil
}

func getAccounts(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	ret := dict.New()
	ctx.State().GetMap(VarStateAllAccounts).Iterate(func(elemKey []byte, val []byte) bool {
		ret.Set(kv.Key(elemKey), val)
		return true
	})
	return codec.NewCodec(ret), nil
}

// deposit moves balances to the sender's account
func deposit(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	if !MoveBetweenAccounts(ctx.AccessState(), ctx.MyAgentID(), ctx.Caller(), ctx.Accounts().Incoming()) {
		return nil, fmt.Errorf("failed to deposit to %s", ctx.Caller().String())
	}
	return nil, nil
}

func move(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	moveTo, ok, err := ctx.Params().GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamsAgentIDNotFound
	}
	if !MoveBetweenAccounts(ctx.AccessState(), ctx.MyAgentID(), *moveTo, ctx.Accounts().Incoming()) {
		return nil, fmt.Errorf("failed to move to %s", moveTo.String())
	}
	return nil, nil
}

func withdraw(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	caller := ctx.Caller()
	if caller.IsAddress() {
		return nil, fmt.Errorf("can't send tokens, must be an address")
	}
	bals, ok := GetAccountBalances(ctx.AccessState(), caller)
	if !ok {
		return nil, fmt.Errorf("withdraw: account not found")
	}
	if !ctx.TransferToAddress(caller.MustAddress(), accounts.NewColoredBalancesFromMap(bals)) {
		return nil, fmt.Errorf("withdraw: account not found")
	}
	return nil, nil
}
