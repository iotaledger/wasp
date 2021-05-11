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

// getBalance returns colored balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func getBalance(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.getBalance")
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	aid, err := params.GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	return getAccountBalanceDict(ctx, getAccountR(ctx.State(), aid), fmt.Sprintf("getBalance for %s", aid)), nil
}

// getTotalAssets returns total colored balances controlled by the chain
func getTotalAssets(ctx coretypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.getTotalAssets")
	return getAccountBalanceDict(ctx, getTotalAssetsAccountR(ctx.State()), "getTotalAssets"), nil
}

// getAccounts returns list of all accounts as keys of the ImmutableCodec
func getAccounts(ctx coretypes.SandboxView) (dict.Dict, error) {
	return getAccountsIntern(ctx.State()), nil
}

// deposit moves transfer to the specified account on the chain
// can be send as request or can be called
// Params:
// - ParamAgentID. default is ctx.Caller(), i.e. deposit on own account
//   in case ParamAgentID. == ctx.Caller() and it is an on-chain call, it means NOP
func deposit(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.deposit.begin -- %s", ctx.IncomingTransfer())

	state := ctx.State()
	mustCheckLedger(state, "accounts.deposit.begin")
	defer mustCheckLedger(state, "accounts.deposit.exit")

	caller := ctx.Caller()
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	targetAgentID := params.MustGetAgentID(ParamAgentID, *caller)
	targetAgentID = commonaccount.Adjust(targetAgentID, ctx.ChainID(), ctx.ChainOwnerID())
	// funds currently are in the owners account, they must be moved to the target
	ownersAccount := coretypes.NewAgentID(ctx.ChainID().AsAddress(), 0)
	succ := MoveBetweenAccounts(state, ownersAccount, targetAgentID, ctx.IncomingTransfer())
	assert.NewAssert(ctx.Log()).Require(succ, "internal error: failed to deposit to %s", caller.String())

	incoming := ctx.IncomingTransfer()
	ctx.Log().Debugf("accounts.deposit.success: target: %s\n%s", targetAgentID, incoming.String())
	return nil, nil
}

// withdraw sends caller's funds to the caller, the address on L1.
// caller must be an address
func withdraw(ctx coretypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	mustCheckLedger(state, "accounts.withdraw.begin")
	defer mustCheckLedger(state, "accounts.withdraw.exit")

	a := assert.NewAssert(ctx.Log())

	a.Require(!ctx.Caller().Address().Equals(ctx.ChainID().AsAddress()), "caller can't be from the same chain")

	account := ctx.Caller()
	account = commonaccount.Adjust(account, ctx.ChainID(), ctx.ChainOwnerID())
	bals, ok := GetAccountBalances(state, account)
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	// sending back to default entry point
	sendTokens := ledgerstate.NewColoredBalances(bals)

	// bring balances to the current account (owner's account)
	a.Require(MoveBetweenAccounts(state, account, commonaccount.Get(ctx.ChainID()), sendTokens),
		"accounts.withdraw.inconsistency. failed to move tokens to owner's account")

	a.Require(ctx.Send(ctx.Caller().Address(), sendTokens, &coretypes.SendMetadata{
		TargetContract: ctx.Caller().Hname(),
	}), "accounts.withdraw.inconsistency: failed sending tokens ")

	ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", sendTokens.String())
	return nil, nil
}
