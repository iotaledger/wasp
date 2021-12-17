package accounts

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
)

var Processor = Contract.Processor(initialize,
	FuncViewBalance.WithHandler(viewBalance),
	FuncViewTotalAssets.WithHandler(viewTotalAssets),
	FuncViewAccounts.WithHandler(viewAccounts),
	FuncDeposit.WithHandler(deposit),
	FuncSendTo.WithHandler(sendTo),
	FuncWithdraw.WithHandler(withdraw),
	FuncHarvest.WithHandler(harvest),
	FuncGetAccountNonce.WithHandler(getAccountNonce),
)

// initialize the init call
func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.initialize.success hname = %s", Contract.Hname().String())
	return nil, nil
}

// viewBalance returns colored balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func viewBalance(ctx iscp.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.viewBalance")
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	aid, err := params.GetAgentID(ParamAgentID)
	if err != nil {
		return nil, err
	}
	return getAccountBalanceDict(getAccountR(ctx.State(), aid)), nil
}

// viewTotalAssets returns total colored balances controlled by the chain
func viewTotalAssets(ctx iscp.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return getAccountBalanceDict(getTotalAssetsAccountR(ctx.State())), nil
}

// viewAccounts returns list of all accounts as keys of the ImmutableCodec
func viewAccounts(ctx iscp.SandboxView) (dict.Dict, error) {
	return getAccountsIntern(ctx.State()), nil
}

// deposit is a generic function to deposit funds to the sender's chain account
func deposit(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.deposit -- %s", ctx.IncomingTransfer())
	// No need to do anything here because funds are already credited to the users account
	// just send back anything that might have been included in "transfer" by mistake
	if ctx.IncomingTransfer() == nil {
		return nil, nil
	}
	sendIncomingTo(ctx, ctx.Caller())
	return nil, nil
}

// sendTo moves transfer to the specified account on the chain. Can be send as request or can be called
// If the target account is a core contract on the same chain
// Params:
// - ParamAgentID. default is ctx.Caller(), i.e. deposit to the own account
//   in case ParamAgentID. == ctx.Caller() and it is an on-chain call, it means NOP
func sendTo(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("accounts.sendTo.begin -- %s", ctx.IncomingTransfer())

	if ctx.IncomingTransfer() == nil {
		return nil, nil
	}
	checkLedger(ctx.State(), "accounts.sendTo.begin")

	caller := ctx.Caller()
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	targetAccount := params.MustGetAgentID(ParamAgentID, caller)

	sendIncomingTo(ctx, targetAccount)
	ctx.Log().Debugf("accounts.sendTo.success: target: %s\n%s",
		targetAccount, ctx.IncomingTransfer().String())

	checkLedger(ctx.State(), "accounts.sendTo.exit")
	return nil, nil
}

func sendIncomingTo(ctx iscp.Sandbox, targetAccount *iscp.AgentID) {
	// funds currently are in the common account (because call is to 'accounts'), they must be moved to the target
	ok := MoveBetweenAccounts(ctx.State(), commonaccount.Get(ctx.ChainID()), targetAccount, ctx.IncomingTransfer())
	assert.NewAssert(ctx.Log()).Require(ok, "internal error: failed to send funds to %s", targetAccount.String())
}

// withdraw sends caller's funds to the caller
func withdraw(ctx iscp.Sandbox) (dict.Dict, error) {
	state := ctx.State()
	checkLedger(state, "accounts.withdraw.begin")

	if ctx.Caller().Address().Equal(ctx.ChainID().AsAddress()) {
		// if the caller is on the same chain, do nothing
		return nil, nil
	}
	// TODO maybe we need to deduct gas fees balance tokensToWithdraw? - to check
	tokensToWithdraw, ok := GetAccountAssets(state, ctx.Caller())
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}
	// will be sending back to default entry point
	a := assert.NewAssert(ctx.Log())
	// the balances included in the transfer (if any) are already in the common account
	// bring all remaining caller's balances to the current account (owner's account).
	// It is needed for subsequent Send call
	a.Require(MoveBetweenAccounts(state, ctx.Caller(), commonaccount.Get(ctx.ChainID()), tokensToWithdraw),
		"accounts.withdraw.inconsistency. failed to move tokens to owner's account")
	// add incoming tokens (after fees) to the balances to be withdrawn.
	// Otherwise, they would end up in the common account
	tokensToWithdraw.Add(ctx.IncomingTransfer())
	// Now all caller's assets are in the common account
	// TODO: by default should be withdrawn all tokens and account should be closed
	//  Introduce "ParamEnsureMinimum = N iotas" which leaves at least N iotas in the account
	// Send call assumes tokens are in the current account
	sendMetadata := &iscp.SendMetadata{
		TargetContract: ctx.Caller().Hname(),
		// other metadata parameters are not important
	}
	a.Require(
		ctx.Send(iscp.RequestParameters{
			TargetAddress: ctx.Caller().Address(),
			Assets:        tokensToWithdraw,
			Metadata:      sendMetadata,
			Options:       nil,
		}),
		"accounts.withdraw.inconsistency: failed sending tokens ",
	)

	ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", tokensToWithdraw.String())

	checkLedger(state, "accounts.withdraw.exit")
	return nil, nil
}

// owner of the chain moves all tokens balance the common account to its own account
// Params:
//   ParamWithdrawAmount if do not exist or is 0 means withdraw all balance
//   ParamWithdrawAssetID assetID to withdraw if amount is specified. Defaults to Iota
func harvest(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireChainOwner(ctx, "harvest")

	state := ctx.State()
	checkLedger(state, "accounts.harvest.begin")
	defer checkLedger(state, "accounts.harvest.exit")

	par := kvdecoder.New(ctx.Params(), ctx.Log())
	// if ParamWithdrawAmount > 0, take it as exact amount to withdraw
	// otherwise assume harvest all
	amount := par.MustGetUint64(ParamWithdrawAmount, 0)

	// default is harvest specified amount of iotas
	assetID := par.MustGetBytes(ParamWithdrawAssetID, iscp.IotaAssetID)

	sourceAccount := commonaccount.Get(ctx.ChainID())
	balance, ok := GetAccountAssets(state, sourceAccount)
	if !ok {
		// empty balance, nothing to withdraw
		return nil, nil
	}

	tokensToSend := iscp.NewEmptyAssets()
	if iscp.IsIota(assetID) {
		if amount > 0 {
			tokensToSend.Iotas = amount
		} else {
			tokensToSend.Iotas = balance.Iotas
		}
	} else {
		token := &iotago.NativeToken{
			ID: iscp.MustNativeTokenIDFromBytes(assetID),
		}
		if amount > 0 {
			token.Amount = new(big.Int).SetUint64(amount)
		} else {
			tokenset, err := balance.Tokens.Set()
			a.RequireNoError(err)
			token.Amount = tokenset[token.ID].Amount
		}
		tokensToSend.Tokens = append(tokensToSend.Tokens, token)
	}

	a.Require(MoveBetweenAccounts(state, sourceAccount, ctx.Caller(), tokensToSend),
		"accounts.harvest.inconsistency. failed to move tokens to owner's account")
	return nil, nil
}

func getAccountNonce(ctx iscp.SandboxView) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	account := par.MustGetAgentID(ParamAgentID)
	nonce := GetMaxAssumedNonce(ctx.State(), account.Address())
	ret := dict.New()
	ret.Set(ParamAccountNonce, codec.EncodeUint64(nonce))
	return ret, nil
}
