package accounts

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

// initialize the init call
func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

// viewBalance returns colored balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func viewBalance(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.viewBalance")
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	aid, err := params.GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	return getAccountBalanceDict(ctx, getAccountR(ctx.State(), aid), fmt.Sprintf("viewBalance for %s", aid)), nil
}

// viewTotalAssets returns total colored balances controlled by the chain
func viewTotalAssets(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return getAccountBalanceDict(ctx, getTotalAssetsAccountR(ctx.State()), "viewTotalAssets"), nil
}

// viewAccounts returns list of all accounts as keys of the ImmutableCodec
func viewAccounts(ctx coretypes.SandboxView) (dict.Dict, error) {
	return getAccountsIntern(ctx.State()), nil
}

// deposit moves transfer to the specified account on the chain. Can be send as request or can be called
// If the target account is a core contract on the same chain, it is adjusted to the common account
// (controlled by the chain owner)
// Params:
// - ParamAgentID. default is ctx.Caller(), i.e. deposit to the own account
//   in case ParamAgentID. == ctx.Caller() and it is an on-chain call, it means NOP
func deposit(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.deposit.begin -- %s", ctx.IncomingTransfer())

	mustCheckLedger(ctx.State(), "accounts.deposit.begin")
	defer mustCheckLedger(ctx.State(), "accounts.deposit.exit")

	caller := ctx.Caller()
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	targetAccount := params.MustGetAgentID(ParamAgentID, *caller)
	targetAccount = commonaccount.AdjustIfNeeded(targetAccount, ctx.ChainID())

	// funds currently are in the common account (because call is to 'accounts'), they must be moved to the target
	succ := MoveBetweenAccounts(ctx.State(), commonaccount.Get(ctx.ChainID()), targetAccount, ctx.IncomingTransfer())
	assert.NewAssert(ctx.Log()).Require(succ, "internal error: failed to deposit to %s", targetAccount.String())

	ctx.Log().Debugf("accounts.deposit.success: target: %s\n%s",
		targetAccount, ctx.IncomingTransfer().String())
	return nil, nil
}

// withdraw sends caller's funds to the caller
func withdraw(ctx coretypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	mustCheckLedger(state, "accounts.withdraw.begin")
	defer mustCheckLedger(state, "accounts.withdraw.exit")

	if ctx.Caller().Address().Equals(ctx.ChainID().AsAddress()) {
		// if the caller is on the same chain, do nothing
		return nil, nil
	}
	bals, ok := GetAccountBalances(state, ctx.Caller())
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	// will be sending back to default entry point

	tokensToSend := ledgerstate.NewColoredBalances(bals)

	a := assert.NewAssert(ctx.Log())
	// bring balances to the current account (owner's account). It is needed for subsequent Send call
	a.Require(MoveBetweenAccounts(state, ctx.Caller(), commonaccount.Get(ctx.ChainID()), tokensToSend),
		"accounts.withdraw.inconsistency. failed to move tokens to owner's account")

	// Send call assumes tokens in the current account
	a.Require(ctx.Send(ctx.Caller().Address(), tokensToSend, &coretypes.SendMetadata{
		TargetContract: ctx.Caller().Hname(),
	}), "accounts.withdraw.inconsistency: failed sending tokens ")

	ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", tokensToSend.String())
	return nil, nil
}

// owner of the chain moves all tokens from the common account to its own account
// Params:
//   ParamWithdrawAmount if do not exist or is 0 means withdraw all balance
//   ParamWithdrawColor color to withdraw if amount is specified. Defaults to ledgerstate.ColorIOTA
func harvest(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireChainOwner(ctx, "harvest")

	state := ctx.State()
	mustCheckLedger(state, "accounts.withdraw.begin")
	defer mustCheckLedger(state, "accounts.withdraw.exit")

	par := kvdecoder.New(ctx.Params(), ctx.Log())
	// if ParamWithdrawAmount > 0, take it as exact amount to withdraw
	// otherwise assume harvest all
	amount, err := par.GetUint64(ParamWithdrawAmount)
	harvestAll := true
	if err == nil && amount > 0 {
		harvestAll = false
	}
	// if color not specified and amount is specified, default is harvest specified amount of iotas
	color := par.MustGetColor(ParamWithdrawColor, ledgerstate.ColorIOTA)

	sourceAccount := commonaccount.Get(ctx.ChainID())
	bals, ok := GetAccountBalances(state, sourceAccount)
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	var tokensToSend *ledgerstate.ColoredBalances
	if harvestAll {
		tokensToSend = ledgerstate.NewColoredBalances(bals)
	} else {
		balCol := bals[color]
		a.Require(balCol >= amount, "accounts.harvest.error: not enough tokens")
		tokensToSend = ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
			color: amount,
		})
	}

	a.Require(MoveBetweenAccounts(state, sourceAccount, ctx.Caller(), tokensToSend),
		"accounts.harvest.inconsistency. failed to move tokens to owner's account")
	return nil, nil
}
