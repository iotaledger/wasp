// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCSandbox::getChainID
func (h *magicContractHandler) GetChainID() isc.ChainID {
	return h.ctx.ChainID()
}

// handler for ISCSandbox::getChainAdmin
func (h *magicContractHandler) GetChainAdmin() iscmagic.ISCAgentID {
	return iscmagic.WrapISCAgentID(h.ctx.ChainAdmin())
}

// handler for ISCSandbox::getTimestampUnixSeconds
func (h *magicContractHandler) GetTimestampUnixSeconds() int64 {
	return h.ctx.Timestamp().Unix()
}

// handler for ISCSandbox::callView
func (h *magicContractHandler) CallView(msg iscmagic.ISCMessage) isc.CallArguments {
	return h.callView(msg.Unwrap())
}

// handler for ISCSandbox::getAllowanceFrom
func (h *magicContractHandler) GetAllowanceFrom(addr common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, addr, h.caller))
}

// handler for ISCSandbox::getAllowanceTo
func (h *magicContractHandler) GetAllowanceTo(target common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, h.caller, target))
}

// handler for ISCSandbox::getAllowance
func (h *magicContractHandler) GetAllowance(from, to common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, from, to))
}

// handler for ISCSandbox::getBaseTokenInfo
func (h *magicContractHandler) GetBaseTokenInfo() iscmagic.IotaCoinInfo {
	return h.GetCoinInfo(iscmagic.CoinType(coin.BaseTokenType.String()))
}

// handler for ISCSandbox::getCoinInfo
func (h *magicContractHandler) GetCoinInfo(coinType iscmagic.CoinType) iscmagic.IotaCoinInfo {
	info, ok := h.ctx.GetCoinInfo(coin.BaseTokenType)
	h.ctx.Requiref(ok, errUnknownCoin.Error())
	return iscmagic.WrapIotaCoinInfo(info)
}

// handler for ISCSandbox::ERC20CoinAddress
func (h *magicContractHandler) ERC20CoinAddress(coinType iscmagic.CoinType) common.Address {
	return iscmagic.ERC20CoinAddress(coin.MustTypeFromString(coinType))
}
