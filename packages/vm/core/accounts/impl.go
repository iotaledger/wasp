package accounts

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
)

func CommonAccount() isc.AgentID {
	return isc.NewAddressAgentID(
		&cryptolib.Address{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	)
}

var Processor = Contract.Processor(nil,
	// funcs
	FuncDeposit.WithHandler(deposit),
	FuncTransferAllowanceTo.WithHandler(transferAllowanceTo),
	FuncWithdraw.WithHandler(withdraw),
	SetCoinMetadata.WithHandler(setCoinMetadata),
	DeleteCoinMetadata.WithHandler(deleteCoinMetadata),
	AdjustCommonAccountBaseTokens.WithHandler(adjustCommonAccountBaseTokens),

	// views
	ViewAccountObjects.WithHandler(viewAccountObjects),
	ViewBalance.WithHandler(viewBalance),
	ViewBalanceBaseToken.WithHandler(viewBalanceBaseToken),
	ViewBalanceBaseTokenEVM.WithHandler(viewBalanceBaseTokenEVM),
	ViewBalanceCoin.WithHandler(viewBalanceCoin),
	ViewGetAccountNonce.WithHandler(viewGetAccountNonce),
	ViewTotalAssets.WithHandler(viewTotalAssets),
)

// SetInitialState initializes the state with provided base tokens and coin information, associating them with the common account.
// this expects the origin amount minus SD
func (s *StateWriter) SetInitialState(baseTokensOnAnchor coin.Value, baseTokenCoinInfo *parameters.IotaCoinInfo) {
	// initial load with base tokens from origin anchor output exceeding minimum storage deposit assumption
	s.SaveCoinInfo(baseTokenCoinInfo)
	s.CreditToAccount(CommonAccount(), isc.NewCoinBalances().Add(coin.BaseTokenType, baseTokensOnAnchor))
}

// deposit is a function to deposit attached assets to the sender's chain account
// It does nothing because assets are already on the sender's account
// Allowance is ignored
func deposit(ctx isc.Sandbox) {
	ctx.Log().Debugf("accounts.deposit: %s", ctx.Request().Assets())
}

// transferAllowanceTo moves whole allowance from the caller to the specified account on the chain.
// Can be sent as a request (sender is the caller) or can be called
// Params:
// - ParamAgentID. AgentID. Required
func transferAllowanceTo(ctx isc.Sandbox, targetAccount isc.AgentID) {
	allowance := ctx.AllowanceAvailable().Clone()
	ctx.TransferAllowedFunds(targetAccount)

	if targetAccount.Kind() != isc.AgentIDKindEthereumAddress {
		return // done
	}
	if !ctx.Caller().Equals(ctx.Request().SenderAccount()) {
		return // only issue "custom EVM tx" when this function is called directly by the request sender
	}
	// issue a "custom EVM tx" so the funds appear on the explorer
	ctx.Call(
		evm.FuncNewL1Deposit.Message(ctx.Caller(), targetAccount.(*isc.EthereumAddressAgentID).EthAddress(), allowance),
		isc.NewEmptyAssets(),
	)
	ctx.Log().Debugf("accounts.transferAllowanceTo.success: target: %s\n%s", targetAccount, ctx.AllowanceAvailable())
}

var errCallerMustHaveL1Address = coreerrors.Register("caller must have L1 address").Create()

// withdraw sends the allowed funds to the caller's L1 address,
func withdraw(ctx isc.Sandbox) {
	allowance := ctx.AllowanceAvailable()
	ctx.Log().Debugf("accounts.withdraw.begin -- %s", allowance)
	if allowance.IsEmpty() {
		panic(ErrNotEnoughAllowance)
	}

	caller := ctx.Caller()
	if _, ok := caller.(*isc.ContractAgentID); ok {
		// cannot withdraw from contract account
		panic(vm.ErrUnauthorized)
	}

	// simple case, caller is not a contract, this is a straightforward withdrawal to L1
	callerAddress, ok := isc.AddressFromAgentID(caller)
	if !ok {
		panic(errCallerMustHaveL1Address)
	}
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
}

func setCoinMetadata(ctx isc.Sandbox, coinInfo *parameters.IotaCoinInfo) {
	ctx.RequireCallerIsChainAdmin()
	NewStateWriterFromSandbox(ctx).SaveCoinInfo(coinInfo)
}

func deleteCoinMetadata(ctx isc.Sandbox, coinType coin.Type) {
	ctx.RequireCallerIsChainAdmin()
	NewStateWriterFromSandbox(ctx).DeleteCoinInfo(coinType)
}

func adjustCommonAccountBaseTokens(ctx isc.Sandbox, credit, debit coin.Value) {
	ctx.RequireCallerIsChainAdmin()
	state := NewStateWriterFromSandbox(ctx)
	if credit != 0 {
		state.CreditToAccount(CommonAccount(), isc.NewCoinBalances().AddBaseTokens(credit))
	}
	if debit != 0 {
		// deduct in order of priority from common account, then payout, then chain admin
		for _, account := range []struct {
			AgentID    isc.AgentID
			MinBalance coin.Value
		}{
			{
				AgentID:    CommonAccount(),
				MinBalance: lo.Must(governance.ViewGetGasCoinTargetValue.DecodeOutput(ctx.CallView(governance.ViewGetGasCoinTargetValue.Message()))),
			},
			{AgentID: lo.Must(governance.ViewGetPayoutAgentID.DecodeOutput(ctx.CallView(governance.ViewGetPayoutAgentID.Message())))},
			{AgentID: ctx.Caller()},
		} {
			if debit == 0 {
				break
			}
			balance := state.GetBaseTokensBalanceDiscardExtraDecimals(account.AgentID)
			ctx.Log().Infof("adjustCommonAccountBaseTokens: must debit %s, trying account %s: balance %s, min %s",
				debit, account.AgentID, balance, account.MinBalance)
			if balance <= account.MinBalance {
				continue
			}
			availableToDeduct := balance - account.MinBalance
			deduct := min(availableToDeduct, debit)
			ctx.Log().Infof("adjustCommonAccountBaseTokens: available to deduct from account %s, deducting %s",
				availableToDeduct, deduct)
			if deduct > 0 {
				state.DebitFromAccount(account.AgentID, isc.NewCoinBalances().AddBaseTokens(deduct))
				debit -= deduct
			}
		}
		if debit > 0 {
			ctx.Log().Debugf("adjustCommonAccountBaseTokens: still remaining debit %s, which is ignored", debit)
		}
	}
}
