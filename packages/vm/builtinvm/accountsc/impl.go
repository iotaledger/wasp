package accountsc

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize the init call
func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accountsc.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

// getBalance returns colored balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func getBalance(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("getBalance")
	aid, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamAgentID))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamWrongOrNotFound
	}

	return getAccountBalanceDict(ctx, getAccount(ctx.State(), aid), fmt.Sprintf("getBalance for %s", aid)), nil
}

// getTotalAssets returns total colored balances controlled by the chain
func getTotalAssets(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("getTotalAssets")
	return getAccountBalanceDict(ctx, getTotalAssetsAccount(ctx.State()), "getTotalAssets"), nil
}

// getAccounts returns list of all accounts as keys of the ImmutableCodec
func getAccounts(ctx vmtypes.SandboxView) (dict.Dict, error) {
	return GetAccounts(ctx.State()), nil
}

// deposit moves transfer to the specified account
// Params:
// - ParamAgentID. default is ctx.Caller()
func deposit(ctx vmtypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	MustCheckLedger(state, "accountsc.deposit.begin")
	defer MustCheckLedger(state, "accountsc.deposit.exit")

	ctx.Eventf("accountsc.deposit.begin -- %s", cbalances.Str(ctx.IncomingTransfer()))
	targetAgentID := ctx.Caller()
	aid, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamAgentID))
	if err != nil {
		return nil, err
	}
	if ok {
		targetAgentID = aid
	}
	// funds currently are at the disposition of accountsc, they are moved to the target
	if !MoveBetweenAccounts(state, ctx.AgentID(), targetAgentID, ctx.IncomingTransfer()) {
		return nil, fmt.Errorf("failed to deposit to %s", ctx.Caller().String())
	}
	ctx.Eventf("accountsc.deposit.success")
	return nil, nil
}

// move moves funds of one color on-chain or cross-chain.
// Parameters:
// - ParamAgentID the target account
// - ParamChainID the target chain. Default is the same chain
// - ParamColor color of the tokens. Default is iota color
// - ParamAmount the amount to move
// - transfer must contain enough iotas for request and node fee (on top of the request token
//   and node fee for this request)
func move(ctx vmtypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	MustCheckLedger(state, "accountsc.move.begin")
	defer MustCheckLedger(state, "accountsc.move.exit")

	ctx.Eventf("accountsc.move.begin")

	params := ctx.Params()
	moveTo, ok, err := codec.DecodeAgentID(params.MustGet(ParamAgentID))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamWrongOrNotFound
	}
	amount, _, err := codec.DecodeInt64(params.MustGet(ParamAmount))
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, ErrParamWrongOrNotFound
	}

	color, ok, err := codec.DecodeColor(params.MustGet(ParamColor))
	if err != nil {
		return nil, err
	}
	if !ok {
		color = balance.ColorIOTA
	}
	targetChain := ctx.ChainID()
	t, ok, err := codec.DecodeChainID(params.MustGet(ParamChainID))
	if err != nil {
		return nil, err
	}
	if ok {
		targetChain = t
	}
	tokensToMove := map[balance.Color]int64{color: amount}
	caller := ctx.Caller()
	if targetChain == ctx.ChainID() {
		// move on-chain
		move := cbalances.NewFromMap(tokensToMove)
		if !MoveBetweenAccounts(state, caller, moveTo, move) {
			return nil, fmt.Errorf("failed to move to %s: %s", moveTo.String(), move.String())
		}
		ctx.Eventf("accountsc.move.success: %s", move.String())
		return nil, nil
	}
	// move to another chain
	// move all tokens to accountsc
	if !MoveBetweenAccounts(state, caller, ctx.AgentID(), cbalances.NewFromMap(tokensToMove)) {
		return nil, fmt.Errorf("accountsc.move.fail: not enough balance")
	}
	// add all incoming tokens from transfer: it must cointain request token + node fee
	ctx.IncomingTransfer().AddToMap(tokensToMove)
	// post a 'deposit' request to the accountsc on the target chain
	if !ctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: coretypes.NewContractID(targetChain, Interface.Hname()),
		EntryPoint:       coretypes.Hn(FuncDeposit),
		Params:           codec.MakeDict(map[string]interface{}{ParamAgentID: moveTo}),
		Transfer:         cbalances.NewFromMap(tokensToMove),
	}) {
		return nil, fmt.Errorf("failed to post request")
	}
	ctx.Eventf("accountsc.withdraw.success")
	return nil, nil
}

// withdraw sends caller's funds to the caller
// The caller must be an address
// TODO not tested
func withdraw(ctx vmtypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	MustCheckLedger(state, "accountsc.withdraw.begin")
	defer MustCheckLedger(state, "accountsc.withdraw.exit")
	caller := ctx.Caller()
	ctx.Eventf("accountsc.withdraw.begin: caller agentID: %s myContractId: %s", caller.String(), ctx.ContractID().String())

	if !caller.IsAddress() {
		return nil, fmt.Errorf("accountsc.withdraw.fail: caller must be an address")
	}
	bals, ok := GetAccountBalances(state, caller)
	if !ok {
		return nil, fmt.Errorf("accountsc.withdraw.fail. Inconsistency 1, empty account")
	}
	send := cbalances.NewFromMap(bals)
	addr := caller.MustAddress()
	if !DebitFromAccount(state, caller, send) {
		return nil, fmt.Errorf("accountsc.withdraw.fail. Inconsistency 2: DebitFromAccount failed")
	}
	if !ctx.TransferToAddress(addr, send) {
		return nil, fmt.Errorf("accountsc.withdraw.fail: TransferToAddress failed")
	}
	// sent to address
	ctx.Eventf("accountsc.withdraw.success. Sent to address %s -- %s", addr.String(), send.String())
	return nil, nil
}

// allow is similar to the ERC-20 allow function.
// TODO not testes
func allow(ctx vmtypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	MustCheckLedger(state, "accountsc.allow.begin")
	defer MustCheckLedger(state, "accountsc.allow.exit")

	agentID, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamAgentID))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamWrongOrNotFound
	}
	amount, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamAgentID))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrParamWrongOrNotFound
	}
	allowances := datatypes.NewMustMap(state, VarStateAllowances)
	if amount <= 0 {
		allowances.DelAt(agentID[:])
		ctx.Eventf("accountsc.allow.success. %s is not allowed to withdraw funds", agentID.String())
	} else {
		allowances.SetAt(agentID[:], util.Uint64To8Bytes(uint64(amount)))
		ctx.Eventf("accountsc.allow.success. Allow %s to withdraw uo to %d", agentID.String(), amount)
	}
	return nil, nil
}
