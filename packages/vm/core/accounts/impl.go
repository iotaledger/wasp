package accounts

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

func CommonAccount() isc.AgentID {
	return isc.NewAgentID(
		&iotago.Ed25519Address{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	)
}

var Processor = Contract.Processor(nil,
	// funcs
	FuncDeposit.WithHandler(deposit),
	FuncFoundryCreateNew.WithHandler(foundryCreateNew),
	FuncFoundryDestroy.WithHandler(foundryDestroy),
	FuncFoundryModifySupply.WithHandler(foundryModifySupply),
	FuncHarvest.WithHandler(harvest),
	FuncTransferAllowanceTo.WithHandler(transferAllowanceTo),
	FuncWithdraw.WithHandler(withdraw),

	// views
	ViewAccountNFTs.WithHandler(viewAccountNFTs),
	ViewAccountNFTAmount.WithHandler(viewAccountNFTAmount),
	ViewAccountNFTsInCollection.WithHandler(viewAccountNFTsInCollection),
	ViewAccountNFTAmountInCollection.WithHandler(viewAccountNFTAmountInCollection),
	ViewAccountFoundries.WithHandler(viewAccountFoundries),
	ViewAccounts.WithHandler(viewAccounts),
	ViewBalance.WithHandler(viewBalance),
	ViewBalanceBaseToken.WithHandler(viewBalanceBaseToken),
	ViewBalanceNativeToken.WithHandler(viewBalanceNativeToken),
	ViewFoundryOutput.WithHandler(viewFoundryOutput),
	ViewGetAccountNonce.WithHandler(viewGetAccountNonce),
	ViewGetNativeTokenIDRegistry.WithHandler(viewGetNativeTokenIDRegistry),
	ViewNFTData.WithHandler(viewNFTData),
	ViewTotalAssets.WithHandler(viewTotalAssets),
)

// this expects the origin amount minus SD
func SetInitialState(state kv.KVStore, baseTokensOnAnchor uint64) {
	// initial load with base tokens from origin anchor output exceeding minimum storage deposit assumption
	CreditToAccount(state, CommonAccount(), isc.NewAssetsBaseTokens(baseTokensOnAnchor))
}

// deposit is a function to deposit attached assets to the sender's chain account
// It does nothing because assets are already on the sender's account
// Allowance is ignored
func deposit(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.deposit")
	return nil
}

// transferAllowanceTo moves whole allowance from the caller to the specified account on the chain.
// Can be sent as a request (sender is the caller) or can be called
// Params:
// - ParamAgentID. AgentID. Required
func transferAllowanceTo(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.transferAllowanceTo.begin -- %s", ctx.AllowanceAvailable())
	targetAccount := ctx.Params().MustGetAgentID(ParamAgentID)
	ctx.TransferAllowedFunds(targetAccount)
	ctx.Log().Debugf("accounts.transferAllowanceTo.success: target: %s\n%s", targetAccount, ctx.AllowanceAvailable())
	return nil
}

// TODO this is just a temporary value, we need to make deposits fee constant across chains.
const ConstDepositFeeTmp = 1 * isc.Million

// withdraw sends the allowed funds to the caller's L1 address,
// or if the caller is a cross-chain contract, to its account.
func withdraw(ctx isc.Sandbox) dict.Dict {
	allowance := ctx.AllowanceAvailable()
	ctx.Requiref(!allowance.IsEmpty(), "Allowance can't be empty in 'accounts.withdraw'")
	if len(allowance.NFTs) > 1 {
		panic(ErrTooManyNFTsInAllowance)
	}

	caller := ctx.Caller()
	callerContract, ok := caller.(*isc.ContractAgentID)
	if !ok || callerContract.Hname().IsNil() {
		// simple case, caller is not a contract, this is a straightforward withdrawal to L1
		callerAddress, ok := isc.AddressFromAgentID(caller)
		ctx.Requiref(ok, "caller must have L1 address")
		remains := ctx.TransferAllowedFunds(ctx.AccountID())
		ctx.Requiref(remains.IsEmpty(), "internal: allowance remains must be empty")
		ctx.Send(isc.RequestParameters{
			TargetAddress: callerAddress,
			Assets:        allowance,
		})
		ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s: %s",
			callerAddress.String(),
			allowance.String(),
		)
		return nil
	}

	// Caller is a contract, and therefore does not have a L1 address to withdraw to.
	// Instead, we will withdraw from its L2 account on the target chain to its L2
	// account on the caller's chain. This requires a second request to be made by
	// the accounts contract on the target chain, that transfers the assets via L1
	// to the caller's chain and requests accounts.TransferAllowanceTo to transfer
	// the assets into the caller's L2 account on the caller's chain.

	// SPECIAL CONSIDERATIONS:
	// 1. The caller contract needs to have enough extra tokens in its account, to send
	// along as part of the withdrawal request, to cover the gas fee GAS1 and sufficient
	// storage deposit SD1 for the withdrawal request to succeed. The storage deposit SD1
	// needs to be added to the withdrawal amount specified in the allowance, so that it
	// can later be returned to the caller. Note that these amounts depend on the fee
	// structure of the target chain, and can be different from the caller's own chain.
	// 2. The caller contract also needs to make sure that sufficient gas fee GAS2 and
	// storage deposit SD2 for the resulting transfer request on its own chain are sent.
	// These amounts depend on the fee structure of the caller's own chain, so it should
	// be easy to figure out the correct amounts. Add GAS2 and SD2 to the withdrawal
	// allowance, to make sure they end up where they are needed, and add the storage
	// deposit SD2 to the withdrawal amount, so that it can be returned to the caller.
	// 3. Any remaining tokens after deduction of the withdrawal amount and gas fees will
	// end up in the L2 account on the caller's chain of the target chain's core accounts
	// contract, since that is the one invoking the transfer request. These tokens will
	// be irretrievably locked up in that account unless the harvest function is amended
	// to transfer these tokens to the chain owner as well.
	// 4. The caller contract needs to set the gas budget for the withdrawal request to
	// GAS1, otherwise the request could cannibalize GAS2 or even SD2 and again cause
	// the assets to be locked up in the L2 account of the core accounts contract.

	// TODO how to know GAS2 within withdrawal request so we can modify
	// the allowance for the transfer request accordingly?
	// TODO tokens could also be locked up in L2 account of core accounts
	// if gas runs out before they are transferred to caller's L2 account
	// So how do we make sure the GAS2 budget is enough

	// if the caller contract is on the same chain the withdrawal would end up
	// in the same L2 account it is taken from, so we do nothing in that case
	if callerContract.ChainID().Equals(ctx.ChainID()) {
		return nil
	}

	// TODO Silly ConstDepositFeeTmp (1M tokens) needs to be replaced with actual GAS2
	gas2 := ConstDepositFeeTmp

	// save the assets to send to the transfer request
	assets := allowance.Clone()
	// deduct the gas budget GAS2 from the allowance, if possible
	ctx.Requiref(allowance.BaseTokens >= gas2, "insufficient base tokens for GAS2")
	allowance.BaseTokens -= gas2

	// warning: this will transfer the assets into the accounts core contract
	// make sure everything transfers out again, or assets will be stuck forever
	remains := ctx.TransferAllowedFunds(ctx.AccountID())
	ctx.Requiref(remains.IsEmpty(), "internal: allowance remains must be empty")

	// send the assets to the caller contract's L2 account on the caller's chain
	ctx.Send(isc.RequestParameters{
		TargetAddress: callerContract.Address(),
		Assets:        assets,
		Metadata: &isc.SendMetadata{
			TargetContract: Contract.Hname(), // core accounts
			EntryPoint:     FuncTransferAllowanceTo.Hname(),
			Allowance:      allowance,
			Params:         dict.Dict{ParamAgentID: codec.EncodeAgentID(callerContract)},
			GasBudget:      gas2,
		},
	})
	ctx.Log().Debugf("accounts.withdraw.success. Sent to contract %s: %s",
		callerContract.String(),
		allowance.String(),
	)
	return nil
}

// harvest moves all the L2 balances of chain common account to chain owner's account
// Params:
//
//	ParamForceMinimumBaseTokens: specify the number of BaseTokens left on the common account will be not less than MinimumBaseTokensOnCommonAccount constant
//
// TODO refactor owner of the chain moves all tokens balance the common account to its own account
func harvest(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	bottomBaseTokens := ctx.Params().MustGetUint64(ParamForceMinimumBaseTokens, MinimumBaseTokensOnCommonAccount)
	if bottomBaseTokens > MinimumBaseTokensOnCommonAccount {
		bottomBaseTokens = MinimumBaseTokensOnCommonAccount
	}

	state := ctx.State()
	commonAccount := CommonAccount()
	toWithdraw := GetAccountFungibleTokens(state, commonAccount)
	if toWithdraw.BaseTokens <= bottomBaseTokens {
		// below minimum, nothing to withdraw
		return nil
	}
	ctx.Requiref(toWithdraw.BaseTokens > bottomBaseTokens, "assertion failed: toWithdraw.BaseTokens > availableBaseTokens")
	toWithdraw.BaseTokens -= bottomBaseTokens
	MustMoveBetweenAccounts(state, commonAccount, ctx.Caller(), toWithdraw)
	return nil
}

// Params:
// - token scheme
// - must be enough allowance for the storage deposit
func foundryCreateNew(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.foundryCreateNew")

	tokenScheme := ctx.Params().MustGetTokenScheme(ParamTokenScheme, &iotago.SimpleTokenScheme{})
	ts := util.MustTokenScheme(tokenScheme)
	ts.MeltedTokens = util.Big0
	ts.MintedTokens = util.Big0

	// create UTXO
	sn, storageDepositConsumed := ctx.Privileged().CreateNewFoundry(tokenScheme, nil)
	ctx.Requiref(storageDepositConsumed > 0, "storage deposit Consumed > 0: assert failed")
	// storage deposit for the foundry is taken from the allowance and removed from L2 ledger
	debitBaseTokensFromAllowance(ctx, storageDepositConsumed)

	// add to the ownership list of the account
	addFoundryToAccount(ctx.State(), ctx.Caller(), sn)

	ret := dict.New()
	ret.Set(ParamFoundrySN, util.Uint32To4Bytes(sn))
	ctx.Event(fmt.Sprintf("Foundry created, serial number = %d", sn))
	return ret
}

// foundryDestroy destroys foundry if that is possible
func foundryDestroy(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.foundryDestroy")
	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	// check if foundry is controlled by the caller
	state := ctx.State()
	caller := ctx.Caller()
	ctx.Requiref(hasFoundry(state, caller, sn), "foundry #%d is not controlled by the caller", sn)

	out, _, _ := GetFoundryOutput(state, sn, ctx.ChainID())
	simpleTokenScheme := util.MustTokenScheme(out.TokenScheme)
	ctx.Requiref(util.IsZeroBigInt(big.NewInt(0).Sub(simpleTokenScheme.MintedTokens, simpleTokenScheme.MeltedTokens)), "can't destroy foundry with positive circulating supply")

	storageDepositReleased := ctx.Privileged().DestroyFoundry(sn)

	deleteFoundryFromAccount(state, caller, sn)
	DeleteFoundryOutput(state, sn)
	// the storage deposit goes to the caller's account
	CreditToAccount(state, caller, &isc.Assets{
		BaseTokens: storageDepositReleased,
	})
	return nil
}

// foundryModifySupply inflates (mints) or shrinks supply of token by the foundry, controlled by the caller
// Params:
// - ParamFoundrySN serial number of the foundry
// - ParamSupplyDeltaAbs absolute delta of the supply as big.Int
// - ParamDestroyTokens true if destroy supply, false (default) if mint new supply
// NOTE: ParamDestroyTokens is needed since `big.Int` `Bytes()` function does not serialize the sign, only the absolute value
func foundryModifySupply(ctx isc.Sandbox) dict.Dict {
	params := ctx.Params()
	sn := params.MustGetUint32(ParamFoundrySN)
	delta := new(big.Int).Abs(params.MustGetBigInt(ParamSupplyDeltaAbs))
	if util.IsZeroBigInt(delta) {
		return nil
	}
	destroy := params.MustGetBool(ParamDestroyTokens, false)
	state := ctx.State()
	caller := ctx.Caller()
	// check if foundry is controlled by the caller
	ctx.Requiref(hasFoundry(state, caller, sn), "foundry #%d is not controlled by the caller", sn)

	out, _, _ := GetFoundryOutput(state, sn, ctx.ChainID())
	nativeTokenID, err := out.NativeTokenID()
	ctx.RequireNoError(err, "internal")

	// accrue change on the caller's account
	// update native tokens on L2 ledger and transit foundry UTXO
	var storageDepositAdjustment int64
	if deltaAssets := isc.NewEmptyAssets().AddNativeTokens(nativeTokenID, delta); destroy {
		// take tokens to destroy from allowance
		accountID := ctx.AccountID()
		ctx.TransferAllowedFunds(accountID,
			isc.NewAssets(0, iotago.NativeTokens{
				&iotago.NativeToken{
					ID:     nativeTokenID,
					Amount: delta,
				},
			}),
		)
		DebitFromAccount(state, accountID, deltaAssets)
		storageDepositAdjustment = ctx.Privileged().ModifyFoundrySupply(sn, delta.Neg(delta))
	} else {
		CreditToAccount(state, caller, deltaAssets)
		storageDepositAdjustment = ctx.Privileged().ModifyFoundrySupply(sn, delta)
	}

	// adjust base tokens on L2 due to the possible change in storage deposit
	switch {
	case storageDepositAdjustment < 0:
		// storage deposit is taken from the allowance of the caller
		debitBaseTokensFromAllowance(ctx, uint64(-storageDepositAdjustment))
	case storageDepositAdjustment > 0:
		// storage deposit is returned to the caller account
		CreditToAccount(state, caller, isc.NewAssetsBaseTokens(uint64(storageDepositAdjustment)))
	}
	return nil
}
