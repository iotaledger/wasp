package accounts

import (
	"math"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
)

var Processor = Contract.Processor(initialize,
	FuncDeposit.WithHandler(deposit),
	FuncTransferAllowanceTo.WithHandler(transferAllowanceTo),
	FuncWithdraw.WithHandler(withdraw),
	FuncHarvest.WithHandler(harvest),
	FuncGetAccountNonce.WithHandler(getAccountNonce),
	FuncGetNativeTokenIDRegistry.WithHandler(viewGetNativeTokenIDRegistry),
	FuncFoundryCreateNew.WithHandler(foundryCreateNew),
	FuncFoundryDestroy.WithHandler(foundryDestroy),
	FuncFoundryModifySupply.WithHandler(foundryModifySupply),
	FuncFoundryOutput.WithHandler(foundryOutput),
	FuncViewBalance.WithHandler(viewBalance),
	FuncViewTotalAssets.WithHandler(viewTotalAssets),
	FuncViewAccounts.WithHandler(viewAccounts),
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

	// initial load with iotas from origin anchor output exceeding minimum dust deposit assumption
	initialLoadIotas := iscp.NewAssets(iotasOnAnchor-dustDepositAssumptions.AnchorOutput, nil)
	CreditToAccount(ctx.State(), commonaccount.Get(ctx.ChainID()), initialLoadIotas)
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
	var NFTID *iotago.NFTID = nil
	if len(allowance.NFTs) > 0 {
		NFTID = allowance.NFTs[0]
	}
	remains := ctx.TransferAllowedFunds(ctx.AccountID())

	// por las dudas
	ctx.Requiref(remains.IsEmpty(), "internal: allowance left after must be empty")

	caller := ctx.Caller()
	isCallerAContract := caller.Hname() != 0

	if isCallerAContract {
		allowance := iscp.NewAllowanceFromAssets(
			iscp.NewAssetsIotas(fundsToWithdraw.Iotas-ConstDepositFeeTmp),
			nil,
		)
		// send funds to a contract on another chain
		tx := iscp.RequestParameters{
			TargetAddress: ctx.Caller().Address(),
			Assets:        fundsToWithdraw,
			NFTID:         NFTID,
			Metadata: &iscp.SendMetadata{
				TargetContract: Contract.Hname(),
				EntryPoint:     FuncTransferAllowanceTo.Hname(),
				Allowance:      allowance,
				Params:         dict.Dict{ParamAgentID: codec.EncodeAgentID(caller)},
				GasBudget:      math.MaxUint64, // TODO This call will fail if not enough gas, and the funds will be lost (credited to this accounts on the target chain)
			},
		}

		ctx.Send(tx)
		ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", ctx.AllowanceAvailable().String())
		return nil
	}
	tx := iscp.RequestParameters{
		TargetAddress: ctx.Caller().Address(),
		Assets:        fundsToWithdraw,
	}
	ctx.Send(tx)
	ctx.Log().Debugf("accounts.withdraw.success. Sent to address %s", ctx.AllowanceAvailable().String())
	return nil
}

// TODO refactor owner of the chain moves all tokens balance the common account to its own account
// Params:
//   ParamForceMinimumIotas specify how may iotas should be left on the common account
//   but not less that MinimumIotasOnCommonAccount constant
func harvest(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner("accounts.harvest")

	state := ctx.State()
	checkLedger(state, "accounts.harvest.begin")
	defer checkLedger(state, "accounts.harvest.exit")

	bottomIotas := ctx.Params().MustGetUint64(ParamForceMinimumIotas, MinimumIotasOnCommonAccount)
	commonAccount := commonaccount.Get(ctx.ChainID())
	toWithdraw := GetAccountAssets(state, commonAccount)
	if toWithdraw.IsEmpty() {
		// empty toWithdraw, nothing to withdraw. Can't be
		return nil
	}
	if toWithdraw.Iotas > bottomIotas {
		toWithdraw.Iotas -= bottomIotas
	}
	MustMoveBetweenAccounts(state, commonAccount, ctx.Caller(), iscp.NewAllowanceFromAssets(toWithdraw, nil))
	return nil
}

// viewBalance returns colored balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func viewBalance(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewBalance")
	aid, err := ctx.Params().GetAgentID(ParamAgentID)
	ctx.RequireNoError(err)
	return getAccountBalanceDict(getAccountR(ctx.State(), aid))
}

// viewTotalAssets returns total colored balances controlled by the chain
func viewTotalAssets(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return getAccountBalanceDict(getTotalL2AssetsAccountR(ctx.State()))
}

// viewAccounts returns list of all accounts as keys of the ImmutableCodec
func viewAccounts(ctx iscp.SandboxView) dict.Dict {
	return getAccountsIntern(ctx.State())
}

func getAccountNonce(ctx iscp.SandboxView) dict.Dict {
	account := ctx.Params().MustGetAgentID(ParamAgentID)
	nonce := GetMaxAssumedNonce(ctx.State(), account.Address())
	ret := dict.New()
	ret.Set(ParamAccountNonce, codec.EncodeUint64(nonce))
	return ret
}

// viewGetNativeTokenIDRegistry returns all native token ID accounted in the chian
func viewGetNativeTokenIDRegistry(ctx iscp.SandboxView) dict.Dict {
	mapping := getNativeTokenOutputMapR(ctx.State())
	ret := dict.New()
	mapping.MustIterate(func(elemKey []byte, value []byte) bool {
		ret.Set(kv.Key(elemKey), []byte{0xFF})
		return true
	})
	return ret
}

// Params:
// - token scheme
// - token tag
// - max supply big integer
// - must be enough allowance for the dust deposit
func foundryCreateNew(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("accounts.foundryCreateNew")

	tokenScheme := ctx.Params().MustGetTokenScheme(ParamTokenScheme, &iotago.SimpleTokenScheme{})
	tokenTag := ctx.Params().MustGetTokenTag(ParamTokenTag, iotago.TokenTag{})
	tokenMaxSupply := ctx.Params().MustGetBigInt(ParamMaxSupply)

	// create UTXO
	sn, dustConsumed := ctx.Privileged().CreateNewFoundry(tokenScheme, tokenTag, tokenMaxSupply, nil)
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
	ctx.Requiref(util.IsZeroBigInt(out.CirculatingSupply), "can't destroy foundry with positive circulating supply")

	dustDepositReleased := ctx.Privileged().DestroyFoundry(sn)

	deleteFoundryFromAccount(getAccountFoundries(ctx.State(), ctx.Caller()), sn)
	DeleteFoundryOutput(ctx.State(), sn)
	// the dust deposit goes to the caller's account
	CreditToAccount(ctx.State(), ctx.Caller(), &iscp.Assets{
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
		CreditToAccount(ctx.State(), ctx.Caller(), iscp.NewAssetsIotas(uint64(dustAdjustment)))
	}
	return nil
}

// foundryOutput takes serial number and returns corresponding foundry output in serialized form
func foundryOutput(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.foundryOutput")

	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	out, _, _ := GetFoundryOutput(ctx.State(), sn, ctx.ChainID())
	ctx.Requiref(out != nil, "foundry #%d does not exist", sn)
	outBin, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
	ctx.RequireNoError(err, "internal: error while serializing foundry output")
	ret := dict.New()
	ret.Set(ParamFoundryOutputBin, outBin)
	return ret
}
