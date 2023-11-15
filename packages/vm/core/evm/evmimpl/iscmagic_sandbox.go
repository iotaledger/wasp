// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCSandbox::getEntropy
func (h *magicContractHandler) GetEntropy() hashing.HashValue {
	return h.ctx.GetEntropy()
}

// handler for ISCSandbox::triggerEvent
func (h *magicContractHandler) TriggerEvent(s string) {
	// TODO adjust triggerevent and all .sol code
	h.ctx.Event("evm.event", []byte(s))
}

// handler for ISCSandbox::getRequestID
func (h *magicContractHandler) GetRequestID() iscmagic.ISCRequestID {
	return iscmagic.WrapISCRequestID(h.ctx.Request().ID())
}

// handler for ISCSandbox::getSenderAccount
func (h *magicContractHandler) GetSenderAccount() iscmagic.ISCAgentID {
	return iscmagic.WrapISCAgentID(h.ctx.Request().SenderAccount())
}

// handler for ISCSandbox::allow
func (h *magicContractHandler) Allow(target common.Address, allowance iscmagic.ISCAssets) {
	addToAllowance(h.ctx, h.caller.Address(), target, allowance.Unwrap())
}

// handler for ISCSandbox::takeAllowedFunds
func (h *magicContractHandler) TakeAllowedFunds(addr common.Address, allowance iscmagic.ISCAssets) {
	taken := subtractFromAllowance(h.ctx, addr, h.caller.Address(), allowance.Unwrap())
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), addr),
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), h.caller.Address()),
		taken,
	)
}

var errInvalidAllowance = coreerrors.Register("allowance must not be greater than sent tokens").Create()

func (h *magicContractHandler) sendViaMetadata(req isc.RequestParameters) {
	h.adjustStorageDeposit(req)

	// make sure that allowance <= sent tokens, so that the target contract does not
	// spend from the common account
	if !req.Assets.Spend(req.Metadata.Allowance) {
		panic(errInvalidAllowance)
	}

	h.moveAssetsToCommonAccount(req.Assets)

	h.ctx.Privileged().SendOnBehalfOf(
		isc.ContractIdentityFromEVMAddress(h.caller.Address()),
		req,
	)
}

// sendViaTxValue enables the transfer of assets by sending a transaction with value attached.
// This will ignore the base token inside `req.Assets.BaseToken` and will use the proper TxValue of the Tx.
// By design, all transactions reaching the EVM will cause a value transfer from the sender to the recipient.
// In our case, this address is 0x1074000[...]. Which means that before *this* function is reached, the value has already
// been moved from the sender to the 0x1074000[...] address.
// The native assets and the NFT still remains inside the senders account as this is Wasp specific.
// Therefore, this function first moves the sent base token from the 0x1074[...] address to the common account
// Then moves the native assets and NFT from the **sender account** to the common account.
// Then moves all three assets to the actual target address on behalf of the sender.
func (h *magicContractHandler) sendViaTxValue(req isc.RequestParameters, callValue *big.Int) {
	adjustedTxValue, remainder := util.EthereumDecimalsToBaseTokenDecimals(callValue, parameters.L1().BaseToken.Decimals)
	if remainder.Sign() != 0 {
		panic(emulator.ErrNonZeroWeiRemainder.Create(callValue, remainder.Uint64()))
	}

	req.Assets = isc.NewAssets(adjustedTxValue, req.Assets.NativeTokens, req.Assets.NFTs...)

	h.adjustStorageDeposit(req)

	if !req.Assets.Spend(req.Metadata.Allowance) {
		panic(errInvalidAllowance)
	}

	evmAddr := isc.NewEthereumAddressAgentID(h.ctx.ChainID(), iscmagic.Address)

	// Move the already transferred base tokens from the 0x1074 address to the common account
	h.ctx.Privileged().MustMoveBetweenAccounts(
		evmAddr,
		h.ctx.AccountID(),
		isc.NewAssetsBaseTokens(req.Assets.BaseTokens),
	)

	// Move the remaining assets (Native assets, NFT) from the sender to the common account
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), h.caller.Address()),
		h.ctx.AccountID(),
		isc.NewAssets(0, req.Assets.NativeTokens, req.Assets.NFTs...),
	)

	// Move all three assets to the actual recipient
	h.ctx.Privileged().SendOnBehalfOf(
		isc.ContractIdentityFromEVMAddress(h.caller.Address()),
		req,
	)
}

func (h *magicContractHandler) handleCallValue(callValue *big.Int) uint64 {
	adjustedTxValue, remainder := util.EthereumDecimalsToBaseTokenDecimals(callValue, parameters.L1().BaseToken.Decimals)
	if remainder.Sign() != 0 {
		panic(emulator.ErrNonZeroWeiRemainder.Create(callValue, remainder.Uint64()))
	}

	evmAddr := isc.NewEthereumAddressAgentID(h.ctx.ChainID(), iscmagic.Address)
	caller := isc.NewEthereumAddressAgentID(h.ctx.ChainID(), h.caller.Address())

	// Move the already transferred base tokens from the 0x1074 address back to the callers account.
	h.ctx.Privileged().MustMoveBetweenAccounts(
		evmAddr,
		caller,
		isc.NewAssetsBaseTokens(adjustedTxValue),
	)

	return adjustedTxValue
}

// handler for ISCSandbox::send
func (h *magicContractHandler) Send(
	targetAddress iscmagic.L1Address,
	assets iscmagic.ISCAssets,
	adjustMinimumStorageDeposit bool,
	metadata iscmagic.ISCSendMetadata,
	sendOptions iscmagic.ISCSendOptions,
) {
	req := isc.RequestParameters{
		TargetAddress:                 targetAddress.MustUnwrap(),
		Assets:                        assets.Unwrap(),
		AdjustToMinimumStorageDeposit: adjustMinimumStorageDeposit,
		Metadata:                      metadata.Unwrap(),
		Options:                       sendOptions.Unwrap(),
	}

	if h.callValue.BitLen() > 0 {
		additionalCallValue := h.handleCallValue(h.callValue)
		req.Assets.BaseTokens += additionalCallValue
	}

	h.sendViaMetadata(req)
}

// handler for ISCSandbox::call
func (h *magicContractHandler) Call(
	contractHname uint32,
	entryPoint uint32,
	params iscmagic.ISCDict,
	allowance iscmagic.ISCAssets,
) iscmagic.ISCDict {
	callRet := h.call(
		isc.Hname(contractHname),
		isc.Hname(entryPoint),
		params.Unwrap(),
		allowance.Unwrap(),
	)
	return iscmagic.WrapISCDict(callRet)
}

var errBaseTokensNotEnoughForStorageDeposit = coreerrors.Register("base tokens (%d) not enough to cover storage deposit (%d)")

func (h *magicContractHandler) adjustStorageDeposit(req isc.RequestParameters) {
	sd := h.ctx.EstimateRequiredStorageDeposit(req)
	if req.Assets.BaseTokens < sd {
		if !req.AdjustToMinimumStorageDeposit {
			panic(errBaseTokensNotEnoughForStorageDeposit.Create(
				req.Assets.BaseTokens,
				sd,
			))
		}
		req.Assets.BaseTokens = sd
	}
}

// moveAssetsToCommonAccount moves the assets from the caller's L2 account to the common
// account before sending to L1
func (h *magicContractHandler) moveAssetsToCommonAccount(assets *isc.Assets) {
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), h.caller.Address()),
		h.ctx.AccountID(),
		assets,
	)
}

// handler for ISCSandbox::registerERC20NativeToken
func (h *magicContractHandler) RegisterERC20NativeToken(foundrySN uint32, name, symbol string, decimals uint8, allowance iscmagic.ISCAssets) {
	h.ctx.Privileged().CallOnBehalfOf(
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), h.caller.Address()),
		evm.Contract.Hname(),
		evm.FuncRegisterERC20NativeToken.Hname(),
		dict.Dict{
			evm.FieldFoundrySN:         codec.EncodeUint32(foundrySN),
			evm.FieldTokenName:         codec.EncodeString(name),
			evm.FieldTokenTickerSymbol: codec.EncodeString(symbol),
			evm.FieldTokenDecimals:     codec.EncodeUint8(decimals),
		},
		allowance.Unwrap(),
	)
}
