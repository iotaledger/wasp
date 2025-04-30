// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCSandbox::getEntropy
func (h *magicContractHandler) GetEntropy() hashing.HashValue {
	return h.ctx.GetEntropy()
}

// handler for ISCSandbox::triggerEvent
func (h *magicContractHandler) TriggerEvent(s string) {
	h.ctx.Event("evm.event", []byte(s))
}

// handler for ISCSandbox::getRequestID
func (h *magicContractHandler) GetRequestID() isc.RequestID {
	return h.ctx.Request().ID()
}

// handler for ISCSandbox::getSenderAccount
func (h *magicContractHandler) GetSenderAccount() iscmagic.ISCAgentID {
	return iscmagic.WrapISCAgentID(h.ctx.Request().SenderAccount())
}

// handler for ISCSandbox::allow
func (h *magicContractHandler) Allow(target common.Address, allowance iscmagic.ISCAssets) {
	addToAllowance(h.ctx, h.caller, target, allowance.Unwrap())
}

// handler for ISCSandbox::takeAllowedFunds
func (h *magicContractHandler) TakeAllowedFunds(addr common.Address, allowance iscmagic.ISCAssets) {
	assets := allowance.Unwrap()
	subtractFromAllowance(h.ctx, addr, h.caller, assets)
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(addr),
		isc.NewEthereumAddressAgentID(h.caller),
		assets,
	)
	// emit ERC20 events for coins
	for _, log := range makeTransferEvents(h.ctx, addr, h.caller, assets) {
		h.evm.StateDB.AddLog(log)
	}
}

func (h *magicContractHandler) handleCallValue(callValue *uint256.Int) coin.Value {
	adjustedTxValue, _ := util.EthereumDecimalsToBaseTokenDecimals(callValue.ToBig(), parameters.BaseTokenDecimals)

	evmAddr := isc.NewEthereumAddressAgentID(iscmagic.Address)
	caller := isc.NewEthereumAddressAgentID(h.caller)

	// Move the already transferred base tokens from the 0x1074 address back to the callers account.
	h.ctx.Privileged().MustMoveBetweenAccounts(
		evmAddr,
		caller,
		isc.NewAssets(adjustedTxValue),
	)

	return adjustedTxValue
}

// handler for ISCSandbox::transferToL1
func (h *magicContractHandler) TransferToL1(
	targetAddress iotago.Address,
	assets iscmagic.ISCAssets,
) {
	req := isc.RequestParameters{
		TargetAddress: cryptolib.NewAddressFromIota(&targetAddress),
		Assets:        assets.Unwrap(),
	}

	// also send any base tokens included as call value
	if h.callValue.BitLen() > 0 {
		additionalCallValue := h.handleCallValue(h.callValue)
		req.Assets.AddBaseTokens(additionalCallValue)
	}

	h.moveAssetsToCommonAccount(req.Assets)

	// emit ERC20 events for coin transfers
	for _, log := range makeTransferEvents(h.ctx, h.caller, common.Address{}, req.Assets) {
		h.evm.StateDB.AddLog(log)
	}

	h.ctx.Privileged().SendOnBehalfOf(
		isc.ContractIdentityFromEVMAddress(h.caller),
		req,
	)
}

// Deprecated: This is included to support calls to the legacy function ISCSandbox::send.
// It is necessary for tracing past blocks.
func (h *magicContractHandler) Send(
	legacyTarget iscmagic.LegacyL1Address,
	legacyAssets iscmagic.LegacyISCAssets,
	_ bool,
	_ iscmagic.LegacyISCSendMetadata,
	_ iscmagic.LegacyISCSendOptions,
) {
	if len(legacyTarget.Data) != 33 {
		panic("cannot decode legacy address")
	}
	var target iotago.Address
	copy(target[:], legacyTarget.Data[1:])

	assets := iscmagic.ISCAssets{}
	if legacyAssets.BaseTokens > 0 {
		assets.Coins = append(assets.Coins, iscmagic.CoinBalance{
			CoinType: iscmagic.CoinType(iotajsonrpc.IotaCoinType),
			Amount:   legacyAssets.BaseTokens,
		})
	}
	for _ = range legacyAssets.NativeTokens {
		panic("cannot send legacy native tokens")
	}
	for _ = range legacyAssets.Nfts {
		panic("cannot send legacy NFTs")
	}
	h.TransferToL1(target, assets)
}

// handler for ISCSandbox::call
func (h *magicContractHandler) Call(
	msg iscmagic.ISCMessage,
	allowance iscmagic.ISCAssets,
) isc.CallArguments {
	return h.call(msg.Unwrap(), allowance.Unwrap())
}

// moveAssetsToCommonAccount moves the assets from the caller's L2 account to the common
// account before sending to L1
func (h *magicContractHandler) moveAssetsToCommonAccount(assets *isc.Assets) {
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(h.caller),
		h.ctx.AccountID(),
		assets,
	)
}
