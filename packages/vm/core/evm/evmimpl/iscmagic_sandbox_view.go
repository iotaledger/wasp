// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
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
	return h.GetCoinInfo(coin.BaseTokenType.String())
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

// handler for ISCSandbox::balanceOf
func (h *magicContractHandler) BalanceOf(account common.Address) *big.Int {
	var agentID isc.AgentID = isc.NewEthereumAddressAgentID(account)
	result := h.ctx.CallView(accounts.ViewBalanceBaseTokenEVM.Message(&agentID))
	balance, err := accounts.ViewBalanceBaseTokenEVM.DecodeOutput(result)
	h.ctx.RequireNoError(err)
	return balance
}

// handler for ISCSandbox::symbol
func (h *magicContractHandler) Symbol() string {
	return "IOTA"
}

// handler for ISCSandbox::decimals
func (h *magicContractHandler) Decimals() uint8 {
	return uint8(parameters.BaseTokenDecimals)
}

// handler for ISCSandbox::supportsInterface
func (h *magicContractHandler) SupportsInterface(interfaceId [4]byte) bool {
	// ERC165 interface ID (XOR of supportsInterface(bytes4))
	erc165InterfaceId := [4]byte{0x01, 0xff, 0xc9, 0xa7}
	// ERC20 interface ID (XOR of function selectors):
	// - totalSupply(): 0x18160ddd
	// - balanceOf(address): 0x70a08231
	// - transfer(address,uint256): 0xa9059cbb
	// - transferFrom(address,address,uint256): 0x23b872dd
	// - approve(address,uint256): 0x095ea7b3
	// - allowance(address,address): 0xdd62ed3e
	// XOR result: 0x36372b07
	erc20InterfaceId := [4]byte{0x36, 0x37, 0x2b, 0x07}

	return interfaceId == erc165InterfaceId || interfaceId == erc20InterfaceId
}
