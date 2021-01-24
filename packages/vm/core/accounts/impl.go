package accounts

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/util"
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
	return GetAccounts(ctx.State()), nil
}

// deposit moves transfer to the specified account on the chain
// can be send as request or can be called
// Params:
// - ParamAgentID. default is ctx.Caller(), i.e. deposit on own account
//   in case ParamAgentID. == ctx.Caller() and it is a call, it means NOP
func deposit(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.deposit.begin -- %s", cbalances.Str(ctx.IncomingTransfer()))

	state := ctx.State()
	MustCheckLedger(state, "accounts.deposit.begin")
	defer MustCheckLedger(state, "accounts.deposit.exit")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	targetAgentID := params.MustGetAgentID(ParamAgentID, ctx.Caller())

	// funds currently are at the disposition of accounts, they are moved to the target
	succ := MoveBetweenAccounts(state, coretypes.NewAgentIDFromContractID(ctx.ContractID()), targetAgentID, ctx.IncomingTransfer())
	coreutil.NewAssert(ctx.Log()).Require(succ,
		"internal error: failed to deposit to %s", ctx.Caller().String())

	ctx.Log().Debugf("accounts.deposit.success: target: %s\n%s", targetAgentID, ctx.IncomingTransfer().String())
	return nil, nil
}

// withdrawToAddress sends caller's funds to the caller, the address on L1.
// caller must be an address
func withdrawToAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	MustCheckLedger(state, "accounts.withdrawToAddress.begin")
	defer MustCheckLedger(state, "accounts.withdrawToAddress.exit")

	a := coreutil.NewAssert(ctx.Log())

	caller := ctx.Caller()
	a.Require(caller.IsAddress(), "caller must be an address")

	bals, ok := GetAccountBalances(state, caller)
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	ctx.Log().Debugf("accounts.withdrawToAddress.begin: caller agentID: %s myContractId: %s",
		caller.String(), ctx.ContractID().String())

	sendTokens := cbalances.NewFromMap(bals)
	addr := caller.MustAddress()

	// remove tokens from the chain ledger
	a.Require(DebitFromAccount(state, caller, sendTokens),
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
	MustCheckLedger(state, "accounts.withdrawToChain.begin")
	defer MustCheckLedger(state, "accounts.withdrawToChain.exit")

	a := coreutil.NewAssert(ctx.Log())

	caller := ctx.Caller()
	ctx.Log().Debugf("accounts.withdrawToChain.begin: caller agentID: %s myContractId: %s",
		caller.String(), ctx.ContractID().String())

	a.Require(!caller.IsAddress(), "caller must be a smart contract")

	bals, ok := GetAccountBalances(state, caller)
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	toWithdraw := cbalances.NewFromMap(bals)
	callerContract := caller.MustContractID()
	if callerContract.ChainID() == ctx.ContractID().ChainID() {
		// no need to move anything on the same chain
		return nil, nil
	}

	// take to tokens here to 'accounts' from the caller
	succ := MoveBetweenAccounts(ctx.State(), caller, coretypes.NewAgentIDFromContractID(ctx.ContractID()), toWithdraw)
	a.Require(succ, "accounts.withdrawToChain.inconsistency to move tokens between accounts")

	// TODO accounts and other core contracts don't need tokens
	//  possible policy: if caller is a core contract, accrue it all to the chain owner
	succ = ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: Interface.ContractID(callerContract.ChainID()),
		EntryPoint:       coretypes.Hn(FuncDeposit),
		Params: codec.MakeDict(map[string]interface{}{
			ParamAgentID: caller,
		}),
		Transfer: toWithdraw,
	})
	a.Require(succ, "accounts.withdrawToChain.inconsistency: failed to post 'deposit' request")
	return nil, nil
}

// allow is similar to the ERC-20 allow function.
// TODO not tested
func allow(ctx coretypes.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	MustCheckLedger(state, "accounts.allow.begin")
	defer MustCheckLedger(state, "accounts.allow.exit")

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
	allowances := collections.NewMap(state, VarStateAllowances)
	if amount <= 0 {
		allowances.MustDelAt(agentID[:])
		ctx.Log().Debugf("accounts.allow.success. %s is not allowed to withdrawToAddress funds", agentID.String())
	} else {
		allowances.MustSetAt(agentID[:], util.Uint64To8Bytes(uint64(amount)))
		ctx.Log().Debugf("accounts.allow.success. Allow %s to withdrawToAddress uo to %d", agentID.String(), amount)
	}
	return nil, nil
}
