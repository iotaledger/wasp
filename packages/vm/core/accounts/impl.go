package accounts

import (
	"math"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
)

var Processor = Contract.Processor(initialize,
	// views
	FuncGetNativeTokenIDRegistry.WithHandler(viewGetNativeTokenIDRegistry),
	FuncViewBalance.WithHandler(viewBalance),
	FuncViewTotalAssets.WithHandler(viewTotalAssets),
	FuncViewAccounts.WithHandler(viewAccounts),
	FuncViewAccountNFTs.WithHandler(viewAccountNFTs),
	FuncViewNFTData.WithHandler(viewNFTData),
	// funcs
	FuncDeposit.WithHandler(deposit),
	FuncTransferAllowanceTo.WithHandler(transferAllowanceTo),
	FuncWithdraw.WithHandler(withdraw),
	FuncHarvest.WithHandler(harvest),
	FuncGetAccountNonce.WithHandler(getAccountNonce),
	FuncFoundryCreateNew.WithHandler(foundryCreateNew),
	FuncFoundryDestroy.WithHandler(foundryDestroy),
	FuncFoundryModifySupply.WithHandler(foundryModifySupply),
	FuncFoundryOutput.WithHandler(foundryOutput),
)

func initialize(ctx iscp.Sandbox) dict.Dict {
	// validating and storing dust deposit assumption constants
	iotasOnAnchor := ctx.StateAnchor().Deposit
	dustAssumptionsBin := ctx.Params().MustGet(ParamDustDepositAssumptionsBin)
	dustDepositAssumptions, err := transaction.DustDepositAssumptionFromBytes(dustAssumptionsBin)
	// checking if assumptions are consistent
	ctx.Requiref(err == nil && iotasOnAnchor >= dustDepositAssumptions.AnchorOutput,
		"accounts.initialize.fail: %v", ErrDustDepositAssumptionsWrong)
	ctx.State().Set(kv.Key(stateVarMinimumDustDepositAssumptionsBin), dustAssumptionsBin)
	// storing hname as a terminal value of the contract's state root.
	// This way we will be able to retrieve commitment to the contract's state
	ctx.State().Set("", ctx.Contract().Bytes())

	// initial load with iotas from origin anchor output exceeding minimum dust deposit assumption
	initialLoadIotas := iscp.NewFungibleTokens(iotasOnAnchor-dustDepositAssumptions.AnchorOutput, nil)
	CreditToAccount(ctx.State(), ctx.ChainID().CommonAccount(), initialLoadIotas)
	return nil
}

// deposit is a function to deposit attached assets to the sender's chain account
// It does nothing because assets are already on the sender's account
// Allowance is ignored
func deposit(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.deposit")
	return nil
}

// transferAllowanceTo moves whole allowance from the caller to the specified account on the chain.
// Can be sent as a request (sender is the caller) or can be called
// Params:
// - ParamAgentID. AgentID. mandatory
// - ParamForceOpenAccount Bool. Optional, default: false
func transferAllowanceTo(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.transferAllowanceTo.begin -- %s", ctx.AllowanceAvailable())

	targetAccount := ctx.Params().MustGetAgentID(ParamAgentID)
	forceOpenAccount := ctx.Params().MustGetBool(ParamForceOpenAccount, false)

	if forceOpenAccount {
		ctx.TransferAllowedFundsForceCreateTarget(targetAccount)
	} else {
		ctx.TransferAllowedFunds(targetAccount)
	}

	ctx.Log().Debugf("accounts.transferAllowanceTo.success: target: %s\n%s", targetAccount, ctx.AllowanceAvailable())
	return nil
}

// TODO this is just a temporary value, we need to make deposits fee constant across chains.
const ConstDepositFeeTmp = uint64(500)

// withdraw sends caller's funds to the caller on-ledger (cross chain)
// The caller explicitly specify the funds to withdraw via the allowance in the request
// Btw: the whole code of entry point is generic, i.e. not specific to the accounts TODO use this feature
func withdraw(ctx iscp.Sandbox) dict.Dict {
	state := ctx.State()
	checkLedger(state, "accounts.withdraw.begin")

	ctx.Requiref(!ctx.AllowanceAvailable().IsEmpty(), "Allowance can't be empty in 'accounts.withdraw'")

	if ctx.Caller().Address().Equal(ctx.ChainID().AsAddress()) {
		// if the caller is on the same chain, do nothing
		return nil
	}
	// move all allowed funds to the account of the current contract context
	// before saving the allowance budget because after the transfer it is mutated
	allowance := ctx.AllowanceAvailable()
	fundsToWithdraw := allowance.Assets
	var nftID *iotago.NFTID
	if len(allowance.NFTs) > 0 {
		if len(allowance.NFTs) > 1 {
			panic(ErrTooManyNFTsInAllowance)
		}
		nftID = &allowance.NFTs[0]
	}
	remains := ctx.TransferAllowedFunds(ctx.AccountID())

	// por las dudas
	ctx.Requiref(remains.IsEmpty(), "internal: allowance left after must be empty")

	caller := ctx.Caller()
	isCallerAContract := caller.Hname() != 0

	if isCallerAContract {
		allowance := iscp.NewAllowanceFungibleTokens(
			iscp.NewTokensIotas(fundsToWithdraw.Iotas - ConstDepositFeeTmp),
		)
		// send funds to a contract on another chain
		tx := iscp.RequestParameters{
			TargetAddress:  ctx.Caller().Address(),
			FungibleTokens: fundsToWithdraw,
			Metadata: &iscp.SendMetadata{
				TargetContract: Contract.Hname(),
				EntryPoint:     FuncTransferAllowanceTo.Hname(),
				Allowance:      allowance,
				Params:         dict.Dict{ParamAgentID: codec.EncodeAgentID(caller)},
				GasBudget:      math.MaxUint64, // TODO This call will fail if not enough gas, and the funds will be lost (credited to this accounts on the target chain)
			},
		}

		if nftID != nil {
			ctx.SendAsNFT(tx, *nftID)
		} else {
			ctx.Send(tx)
		}
		ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", ctx.AllowanceAvailable().String())
		return nil
	}
	tx := iscp.RequestParameters{
		TargetAddress:  ctx.Caller().Address(),
		FungibleTokens: fundsToWithdraw,
	}
	if nftID != nil {
		ctx.SendAsNFT(tx, *nftID)
	} else {
		ctx.Send(tx)
	}
	ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", ctx.AllowanceAvailable().String())
	return nil
}

// TODO refactor owner of the chain moves all tokens balance the common account to its own account
// Params:
//   ParamForceMinimumIotas specify how may iotas should be left on the common account
//   but not less that MinimumIotasOnCommonAccount constant
func harvest(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	state := ctx.State()
	checkLedger(state, "accounts.harvest.begin")
	defer checkLedger(state, "accounts.harvest.exit")

	bottomIotas := ctx.Params().MustGetUint64(ParamForceMinimumIotas, MinimumIotasOnCommonAccount)
	commonAccount := ctx.ChainID().CommonAccount()
	toWithdraw := GetAccountAssets(state, commonAccount)
	if toWithdraw.IsEmpty() {
		// empty toWithdraw, nothing to withdraw. Can't be
		return nil
	}
	if toWithdraw.Iotas > bottomIotas {
		toWithdraw.Iotas -= bottomIotas
	}
	MustMoveBetweenAccounts(state, commonAccount, ctx.Caller(), toWithdraw, nil)
	return nil
}

// Params:
// - token scheme
// - token tag
// - max supply big integer
// - must be enough allowance for the dust deposit
func foundryCreateNew(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.foundryCreateNew")

	tokenScheme := ctx.Params().MustGetTokenScheme(ParamTokenScheme, &iotago.SimpleTokenScheme{})
	ts := util.MustTokenScheme(tokenScheme)
	ts.MeltedTokens = util.Big0
	ts.MintedTokens = util.Big0
	tokenTag := ctx.Params().MustGetTokenTag(ParamTokenTag, iotago.TokenTag{})

	// create UTXO
	sn, dustConsumed := ctx.Privileged().CreateNewFoundry(tokenScheme, tokenTag, nil)
	ctx.Requiref(dustConsumed > 0, "dustConsumed > 0: assert failed")
	// dust deposit for the foundry is taken from the allowance and removed from L2 ledger
	debitIotasFromAllowance(ctx, dustConsumed)

	// add to the ownership list of the account
	AddFoundryToAccount(ctx.State(), ctx.Caller(), sn)

	ret := dict.New()
	ret.Set(ParamFoundrySN, util.Uint32To4Bytes(sn))
	return ret
}

// foundryDestroy destroys foundry if that is possible
func foundryDestroy(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.foundryDestroy")
	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	// check if foundry is controlled by the caller
	ctx.Requiref(HasFoundry(ctx.State(), ctx.Caller(), sn), "foundry #%d is not controlled by the caller", sn)

	out, _, _ := GetFoundryOutput(ctx.State(), sn, ctx.ChainID())
	simpleTokenScheme := util.MustTokenScheme(out.TokenScheme)
	ctx.Requiref(util.IsZeroBigInt(big.NewInt(0).Sub(simpleTokenScheme.MintedTokens, simpleTokenScheme.MeltedTokens)), "can't destroy foundry with positive circulating supply")

	dustDepositReleased := ctx.Privileged().DestroyFoundry(sn)

	deleteFoundryFromAccount(getAccountFoundries(ctx.State(), ctx.Caller()), sn)
	DeleteFoundryOutput(ctx.State(), sn)
	// the dust deposit goes to the caller's account
	CreditToAccount(ctx.State(), ctx.Caller(), &iscp.FungibleTokens{
		Iotas: dustDepositReleased,
	})
	return nil
}

// foundryModifySupply inflates (mints) or shrinks supply of token by the foundry, controlled by the caller
// Params:
// - ParamFoundrySN serial number of the foundry
// - ParamSupplyDeltaAbs absolute delta of the supply as big.Int
// - ParamDestroyTokens true if destroy supply, false (default) if mint new supply
func foundryModifySupply(ctx iscp.Sandbox) dict.Dict {
	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	delta := ctx.Params().MustGetBigInt(ParamSupplyDeltaAbs)
	if util.IsZeroBigInt(delta) {
		return nil
	}
	destroy := ctx.Params().MustGetBool(ParamDestroyTokens, false)
	// check if foundry is controlled by the caller
	ctx.Requiref(HasFoundry(ctx.State(), ctx.Caller(), sn), "foundry #%d is not controlled by the caller", sn)

	out, _, _ := GetFoundryOutput(ctx.State(), sn, ctx.ChainID())
	tokenID, err := out.NativeTokenID()
	ctx.RequireNoError(err, "internal")

	// accrue change on the caller's account
	// update native tokens on L2 ledger and transit foundry UTXO
	var dustAdjustment int64
	if deltaAssets := iscp.NewEmptyAssets().AddNativeTokens(tokenID, delta); destroy {
		DebitFromAccount(ctx.State(), ctx.Caller(), deltaAssets)
		dustAdjustment = ctx.Privileged().ModifyFoundrySupply(sn, delta.Neg(delta))
	} else {
		CreditToAccount(ctx.State(), ctx.Caller(), deltaAssets)
		dustAdjustment = ctx.Privileged().ModifyFoundrySupply(sn, delta)
	}

	// adjust iotas on L2 due to the possible change in dust deposit
	switch {
	case dustAdjustment < 0:
		// dust deposit is taken from the allowance of the caller
		debitIotasFromAllowance(ctx, uint64(-dustAdjustment))
	case dustAdjustment > 0:
		// dust deposit is returned to the caller account
		CreditToAccount(ctx.State(), ctx.Caller(), iscp.NewTokensIotas(uint64(dustAdjustment)))
	}
	return nil
}
