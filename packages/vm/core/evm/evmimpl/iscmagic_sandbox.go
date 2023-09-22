// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
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
	// 	id := nftID.Unwrap()
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
