// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCSandbox::getChainID
func (h *magicContractHandler) GetChainID() iscmagic.ISCChainID {
	return iscmagic.WrapISCChainID(h.ctx.ChainID())
}

// handler for ISCSandbox::getChainOwnerID
func (h *magicContractHandler) GetChainOwnerID() iscmagic.ISCAgentID {
	return iscmagic.WrapISCAgentID(h.ctx.ChainOwnerID())
}

// handler for ISCSandbox::getNFTData
func (h *magicContractHandler) GetObjectData(objectID iscmagic.ObjectID) iscmagic.ISCNFT {
	nft := h.ctx.GetObjectData(objectID.Unwrap())
	return iscmagic.WrapISCNFT(nft)
}

// handler for ISCSandbox::getIRC27NFTData
func (h *magicContractHandler) GetIRC27NFTData(objectID iscmagic.ObjectID) iscmagic.IRC27NFT {
	nft := h.ctx.GetObjectData(objectID.Unwrap())
	metadata, err := isc.IRC27NFTMetadataFromBytes(nft.Metadata)
	h.ctx.RequireNoError(err)
	return iscmagic.IRC27NFT{
		Nft:      iscmagic.WrapISCNFT(nft),
		Metadata: iscmagic.WrapIRC27NFTMetadata(metadata),
	}
}

// handler for ISCSandbox::getIRC27TokenURI
func (h *magicContractHandler) GetIRC27TokenURI(objectID iscmagic.ObjectID) string {
	nft := h.ctx.GetObjectData(objectID.Unwrap())
	metadata, err := isc.IRC27NFTMetadataFromBytes(nft.Metadata)
	h.ctx.RequireNoError(err)
	return evm.EncodePackedNFTURI(metadata)
}

// handler for ISCSandbox::getTimestampUnixSeconds
func (h *magicContractHandler) GetTimestampUnixSeconds() int64 {
	return h.ctx.Timestamp().Unix()
}

// handler for ISCSandbox::callView
func (h *magicContractHandler) CallView(
	contractHname uint32,
	entryPoint uint32,
	params iscmagic.CallArguments,
) iscmagic.CallArguments {
	callRet := h.callView(isc.NewMessage(
		isc.Hname(contractHname),
		isc.Hname(entryPoint),
		params.Unwrap(),
	))
	return iscmagic.WrapCallArguments(callRet)
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

// handler for ISCSandbox::getBaseTokenProperties
func (h *magicContractHandler) GetBaseTokenProperties() iscmagic.ISCTokenProperties {
	l1 := parameters.L1()
	return iscmagic.ISCTokenProperties{
		Name:         l1.BaseToken.Name,
		TickerSymbol: l1.BaseToken.TickerSymbol,
		Decimals:     uint8(l1.BaseToken.Decimals),
		TotalSupply:  big.NewInt(int64(l1.Protocol.TokenSupply)),
	}
}

// handler for ISCSandbox::erc20NativeTokensAddress
func (h *magicContractHandler) Erc20NativeTokensAddress(foundrySN uint32) common.Address {
	return iscmagic.ERC20NativeTokensAddress(foundrySN)
}

// handler for ISCSandbox::erc721NFTCollectionAddress
func (h *magicContractHandler) Erc721NFTCollectionAddress(collectionID iscmagic.ObjectID) common.Address {
	return iscmagic.ERC721NFTCollectionAddress(collectionID.Unwrap())
}

// handler for ISCSandbox::erc20NativeTokensFoundrySerialNumber
func (h *magicContractHandler) Erc20NativeTokensFoundrySerialNumber(addr common.Address) uint32 {
	sn, err := iscmagic.ERC20NativeTokensFoundrySN(addr)
	h.ctx.RequireNoError(err)
	return sn
}

// handler for ISCSandbox::getNativeTokenID
func (h *magicContractHandler) GetNativeTokenID(foundrySN uint32) iscmagic.CoinType {
	panic("refactor me: GetNativeTokenID evm (is this even still required?)")
	/*r := h.callView(accounts.ViewNativeToken.Message(foundrySN))
	out, err := accounts.ViewNativeToken.Output.Decode(r)
	h.ctx.RequireNoError(err)
	nativeTokenID := out.(*iotago.FoundryOutput).MustNativeTokenID()
	return iscmagic.WrapCoinType(nativeTokenID)*/
}

var errUnsupportedTokenScheme = coreerrors.Register("unsupported TokenScheme kind").Create()

// handler for ISCSandbox::getNativeTokenScheme
func (h *magicContractHandler) GetNativeTokenScheme(foundrySN uint32) iotago.SimpleTokenScheme {
	panic("refactor me: GetNativeTokenScheme evm (is this even still required?)")
	/*r := h.callView(accounts.ViewNativeToken.Message(foundrySN))
	out, err := accounts.ViewNativeToken.Output.Decode(r)
	h.ctx.RequireNoError(err)
	s, ok := out.(*iotago.FoundryOutput).TokenScheme.(*iotago.SimpleTokenScheme)
	if !ok {
		panic(errUnsupportedTokenScheme)
	}
	return *s*/
}
