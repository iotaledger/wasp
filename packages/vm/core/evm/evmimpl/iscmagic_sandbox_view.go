// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCSandbox::getChainID
func (h *magicContractHandler) GetChainID() isc.ChainID {
	return h.ctx.ChainID()
}

// handler for ISCSandbox::getChainOwnerID
func (h *magicContractHandler) GetChainOwnerID() iscmagic.ISCAgentID {
	return iscmagic.WrapISCAgentID(h.ctx.ChainOwnerID())
}

// handler for ISCSandbox::getObjectBCS
func (h *magicContractHandler) GetObjectBCS(objectID sui.ObjectID) []byte {
	bcs, ok := h.ctx.GetObjectBCS(objectID)
	h.ctx.Requiref(ok, errUnknownObject.Error())
	return bcs
}

// handler for ISCSandbox::getIRC27NFTData
func (h *magicContractHandler) GetIRC27NFTData(nftID sui.ObjectID) iscmagic.IRC27NFTMetadata {
	bcs := h.GetObjectBCS(nftID)
	metadata, err := isc.IRC27NFTMetadataFromBCS(bcs)
	h.ctx.RequireNoError(err)
	return iscmagic.WrapIRC27NFTMetadata(metadata)
}

// handler for ISCSandbox::getIRC27TokenURI
func (h *magicContractHandler) GetIRC27TokenURI(nftID sui.ObjectID) string {
	bcs := h.GetObjectBCS(nftID)
	metadata, err := isc.IRC27NFTMetadataFromBCS(bcs)
	h.ctx.RequireNoError(err)
	return evm.EncodePackedNFTURI(metadata)
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
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, addr, h.caller.Address()))
}

// handler for ISCSandbox::getAllowanceTo
func (h *magicContractHandler) GetAllowanceTo(target common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, h.caller.Address(), target))
}

// handler for ISCSandbox::getAllowance
func (h *magicContractHandler) GetAllowance(from, to common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, from, to))
}

// handler for ISCSandbox::getBaseTokenInfo
func (h *magicContractHandler) GetBaseTokenInfo() *isc.SuiCoinInfo {
	return h.GetCoinInfo(coin.BaseTokenType)
}

// handler for ISCSandbox::getCoinInfo
func (h *magicContractHandler) GetCoinInfo(coinType coin.Type) *isc.SuiCoinInfo {
	info, ok := h.ctx.GetCoinInfo(coin.BaseTokenType)
	h.ctx.Requiref(ok, errUnknownCoin.Error())
	return info
}

// handler for ISCSandbox::ERC20CoinAddress
func (h *magicContractHandler) ERC20CoinAddress(coinType coin.Type) common.Address {
	return iscmagic.ERC20CoinAddress(coinType)
}

// handler for ISCSandbox::erc721NFTCollectionAddress
func (h *magicContractHandler) Erc721NFTCollectionAddress(collectionID sui.ObjectID) common.Address {
	return iscmagic.ERC721NFTCollectionAddress(collectionID)
}
