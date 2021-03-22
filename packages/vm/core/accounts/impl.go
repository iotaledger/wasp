package accounts

import (
	"fmt"
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
	targetAgentID = adjustAccount(ctx, targetAgentID)
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
	account = adjustAccount(ctx, account)
	bals, ok := GetAccountBalances(state, account)
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	// sending beck to default entry point
	sendTokens := ledgerstate.NewColoredBalances(bals)
	a.Require(DebitFromAccount(state, account, sendTokens),
		"accounts.withdraw.inconsistency. failed to remove tokens from the chain")

	a.Require(ctx.Send(ctx.Caller().Address(), sendTokens, &coretypes.SendMetadata{
		TargetContract: ctx.Caller().Hname(),
	}), "accounts.withdraw.inconsistency: failed sending tokens ")

	ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", sendTokens.String())
	return nil, nil
}

func adjustAccount(ctx coretypes.Sandbox, agentID *coretypes.AgentID) *coretypes.AgentID {
	if agentID.Equals(ctx.ChainOwnerID()) {
		return ownersAccount(ctx)
	}
	if !agentID.Address().Equals(ctx.ChainID().AsAddress()) {
		return agentID
	}
	switch agentID.Hname() {
	case coretypes.HnameRoot, coretypes.HnameAccounts, coretypes.HnameBlob, coretypes.HnameEventlog:
		return ownersAccount(ctx)
	}
	return agentID
}

func ownersAccount(ctx coretypes.Sandbox) *coretypes.AgentID {
	return coretypes.NewAgentID(ctx.ChainID().AsAddress(), 0)
}
