package accounts

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
	return getAccountBalanceDict(ctx, getAccountR(ctx.State(), aid), fmt.Sprintf("getBalance for %s", &aid)), nil
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

	// funds currently are at the disposition of accounts, they are moved to the target
	from := coretypes.NewAgentIDFromContractID(ctx.ContractID())
	succ := MoveBetweenAccounts(state, from, targetAgentID, ctx.IncomingTransfer())
	assert.NewAssert(ctx.Log()).Require(succ, "internal error: failed to deposit to %s", caller.String())

	incoming := ctx.IncomingTransfer()
	ctx.Log().Debugf("accounts.deposit.success: target: %s\n%s", targetAgentID, incoming.String())
	return nil, nil
}

// withdrawToAddress sends caller's funds to the caller, the address on L1.
// caller must be an address
func withdrawToAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	mustCheckLedger(state, "accounts.withdrawToAddress.begin")
	defer mustCheckLedger(state, "accounts.withdrawToAddress.exit")

	a := assert.NewAssert(ctx.Log())

	a.Require(ctx.Caller().IsContract(), "caller must be an address")

	bals, ok := GetAccountBalances(state, ctx.Caller())
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	cid := ctx.ContractID()
	ctx.Log().Debugf("accounts.withdrawToAddress.begin: caller agentID: %s myContractId: %s",
		ctx.Caller().String(), cid.String())

	sendTokens := ledgerstate.NewColoredBalances(bals)
	addr := ctx.Caller().AsAddress()

	// remove tokens from the chain ledger
	a.Require(DebitFromAccount(state, ctx.Caller(), sendTokens),
		"accounts.withdrawToAddress.inconsistency. failed to remove tokens from the chain")
	// send tokens to address
	a.Require(ctx.TransferToAddress(addr, sendTokens),
		"accounts.withdrawToAddress.inconsistency: failed to transfer tokens to address")

	ctx.Log().Debugf("accounts.withdrawToAddress.success. Sent to address %s -- %s",
		addr.String(), sendTokens.String())
	return nil, nil
}

// withdrawToChain sends caller's funds to the caller via account::deposit.
func withdrawToChain(ctx coretypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	mustCheckLedger(state, "accounts.withdrawToChain.begin")
	defer mustCheckLedger(state, "accounts.withdrawToChain.exit")

	a := assert.NewAssert(ctx.Log())

	caller := ctx.Caller()
	cid := ctx.ContractID()
	ctx.Log().Debugf("accounts.withdrawToChain.begin: caller agentID: %s myContractId: %s",
		caller.String(), cid.String())

	a.Require(caller.IsContract(), "caller must be a smart contract")

	bals, ok := GetAccountBalances(state, caller)
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	toWithdraw := ledgerstate.NewColoredBalances(bals)
	callerContract := caller.MustContractID()
	if callerContract.ChainID() == ctx.ContractID().ChainID() {
		// no need to move anything on the same chain
		return nil, nil
	}

	// take to tokens here to 'accounts' from the caller
	toAgentId := coretypes.NewAgentIDFromContractID(ctx.ContractID())
	succ := MoveBetweenAccounts(ctx.State(), caller, toAgentId, toWithdraw)
	a.Require(succ, "accounts.withdrawToChain.inconsistency to move tokens between accounts")

	succ = ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: Interface.ContractID(*callerContract.ChainID()),
		EntryPoint:       coretypes.Hn(FuncDeposit),
		Params: codec.MakeDict(map[string]interface{}{
			ParamAgentID: caller,
		}),
		Transfer: toWithdraw,
	})
	a.Require(succ, "accounts.withdrawToChain.inconsistency: failed to post 'deposit' request")
	return nil, nil
}
