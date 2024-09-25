package accounts

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func CommonAccount() isc.AgentID {
	return isc.NewAddressAgentID(
		&cryptolib.Address{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	)
}

var Processor = Contract.Processor(nil,
	// funcs
	FuncDeposit.WithHandler(deposit),
	FuncTransferAccountToChain.WithHandler(transferAccountToChain),
	FuncTransferAllowanceTo.WithHandler(transferAllowanceTo),
	FuncWithdraw.WithHandler(withdraw),
	SetCoinMetadata.WithHandler(setCoinMetadata),
	DeleteCoinMetadata.WithHandler(deleteCoinMetadata),

	// views
	ViewAccountObjects.WithHandler(viewAccountObjects),
	ViewAccountObjectsInCollection.WithHandler(viewAccountObjectsInCollection),
	ViewBalance.WithHandler(viewBalance),
	ViewBalanceBaseToken.WithHandler(viewBalanceBaseToken),
	ViewBalanceBaseTokenEVM.WithHandler(viewBalanceBaseTokenEVM),
	ViewBalanceCoin.WithHandler(viewBalanceCoin),
	ViewGetAccountNonce.WithHandler(viewGetAccountNonce),
	ViewObjectBCS.WithHandler(viewObjectBCS),
	ViewTotalAssets.WithHandler(viewTotalAssets),
)

// this expects the origin amount minus SD
func (s *StateWriter) SetInitialState(baseTokensOnAnchor coin.Value, baseTokenCoinInfo *isc.SuiCoinInfo) {
	// initial load with base tokens from origin anchor output exceeding minimum storage deposit assumption
	s.SaveCoinInfo(baseTokenCoinInfo)
	s.CreditToAccount(CommonAccount(), isc.NewCoinBalances().Add(coin.BaseTokenType, baseTokensOnAnchor), isc.ChainID{})
}

// deposit is a function to deposit attached assets to the sender's chain account
// It does nothing because assets are already on the sender's account
// Allowance is ignored
func deposit(ctx isc.Sandbox) {
	ctx.Log().Debugf("accounts.deposit")
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
		nil,
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

func setCoinMetadata(ctx isc.Sandbox, coinInfo *isc.SuiCoinInfo) {
	ctx.RequireCallerIsChainOwner()
	NewStateWriterFromSandbox(ctx).SaveCoinInfo(coinInfo)
}

func deleteCoinMetadata(ctx isc.Sandbox, coinType coin.Type) {
	ctx.RequireCallerIsChainOwner()
	NewStateWriterFromSandbox(ctx).DeleteCoinInfo(coinType)
}

// transferAccountToChain transfers the specified allowance from the sender SC's L2
// account on the target chain to the sender SC's L2 account on the origin chain.
//
// Caller must be a contract, and we will transfer the allowance from its L2 account
// on the target chain to its L2 account on the origin chain. This requires that
// this function takes the allowance into custody and in turn sends the assets as
// allowance to the origin chain, where that chain's accounts.TransferAllowanceTo()
// function then transfers it into the caller's L2 account on that chain.
//
// IMPORTANT CONSIDERATIONS:
// 1. The caller contract needs to provide sufficient base tokens in its
// allowance, to cover the gas fee GAS1 for this request.
// Note that this amount depend on the fee structure of the target chain,
// which can be different from the fee structure of the caller's own chain.
//
// 2. The caller contract also needs to provide sufficient base tokens in
// its allowance, to cover the gas fee GAS2 for the resulting request to
// accounts.TransferAllowanceTo() on the origin chain. The caller needs to
// specify this GAS2 amount through the GasReserve parameter.
//
// 3. The caller contract also needs to provide a storage deposit SD with
// this request, holding enough base tokens *independent* of the GAS1 and
// GAS2 amounts.
// Since this storage deposit is dictated by L1 we can use this amount as
// storage deposit for the resulting accounts.TransferAllowanceTo() request,
// where it will be then returned to the caller as part of the transfer.
//
// 4. This means that the caller contract needs to provide at least
// GAS1 + GAS2 + SD base tokens as assets to this request, and provide an
// allowance to the request that is exactly GAS2 + SD + transfer amount.
// Failure to meet these conditions may result in a failed request and
// worst case the assets sent to accounts.TransferAllowanceTo() could be
// irretrievably locked up in an account on the origin chain that belongs
// to the accounts core contract of the target chain.
//
// 5. The caller contract needs to set the gas budget for this request to
// GAS1 to guard against unanticipated changes in the fee structure that
// raise the gas price, otherwise the request could accidentally cannibalize
// GAS2 or even SD, with potential failure and locked up assets as a result.
func transferAccountToChain(ctx isc.Sandbox, optionalGasReserve *uint64) {
	allowance := ctx.AllowanceAvailable()
	ctx.Log().Debugf("accounts.transferAccountToChain.begin -- %s", allowance)
	if allowance.IsEmpty() {
		panic(ErrNotEnoughAllowance)
	}

	caller := ctx.Caller()
	callerContract, ok := caller.(*isc.ContractAgentID)
	if !ok || callerContract.Hname().IsNil() {
		// caller must be contract
		panic(vm.ErrUnauthorized)
	}

	// if the caller contract is on the same chain the transfer would end up
	// in the same L2 account it is taken from, so we do nothing in that case
	if callerContract.ChainID().Equals(ctx.ChainID()) {
		return
	}

	// save the assets to send to the transfer request, as specified by the allowance
	assets := allowance.Clone()

	// deduct the gas reserve GAS2 from the allowance, if possible
	gasReserve := coreutil.FromOptional(optionalGasReserve, gas.LimitsDefault.MinGasPerRequest)
	gasReserveTokens := ctx.ChainInfo().GasFeePolicy.FeeFromGas(gasReserve, nil, 0)
	if allowance.BaseTokens() < gasReserveTokens {
		panic(ErrNotEnoughAllowance)
	}
	allowance.Coins.Sub(coin.BaseTokenType, gasReserveTokens)

	// Warning: this will transfer all assets into the accounts core contract's L2 account.
	// Be sure everything transfers out again, or assets will be stuck forever.
	ctx.TransferAllowedFunds(ctx.AccountID())

	// Send the specified assets, which should include GAS2 and SD, as part of the
	// accounts.TransferAllowanceTo() request on the origin chain.
	// Note that the assets initially end up in the L2 account of this core accounts
	// contract on the origin chain, from where an allowance of SD plus transfer amount
	// will finally end up in the caller's L2 account on the origin chain.
	ctx.Send(isc.RequestParameters{
		TargetAddress: callerContract.Address(),
		Assets:        assets,
		Metadata: &isc.SendMetadata{
			Message:   FuncTransferAllowanceTo.Message(callerContract),
			Allowance: allowance,
			GasBudget: gasReserve,
		},
	})
	ctx.Log().Debugf("accounts.transferAccountToChain.success. Sent to contract %s: %s",
		callerContract.String(),
		allowance.String(),
	)
}
