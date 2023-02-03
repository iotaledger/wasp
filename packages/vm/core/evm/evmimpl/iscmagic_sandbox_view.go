// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCSandbox::getChainID
func (h *magicContractViewHandler) GetChainID() iscmagic.ISCChainID {
	return iscmagic.WrapISCChainID(h.ctx.ChainID())
}

// handler for ISCSandbox::getChainOwnerID
func (h *magicContractViewHandler) GetChainOwnerID() iscmagic.ISCAgentID {
	return iscmagic.WrapISCAgentID(h.ctx.ChainOwnerID())
}

// handler for ISCSandbox::getNFTData
func (h *magicContractViewHandler) GetNFTData(nftID iscmagic.NFTID) iscmagic.ISCNFT {
	nft := h.ctx.GetNFTData(nftID.Unwrap())
	return iscmagic.WrapISCNFT(nft)
}

// handler for ISCSandbox::getIRC27NFTData
func (h *magicContractViewHandler) GetIRC27NFTData(nftID iscmagic.NFTID) iscmagic.IRC27NFT {
	nft := h.ctx.GetNFTData(nftID.Unwrap())
	metadata, err := transaction.IRC27NFTMetadataFromBytes(nft.Metadata)
	h.ctx.RequireNoError(err)
	return iscmagic.IRC27NFT{
		Nft:      iscmagic.WrapISCNFT(nft),
		Metadata: iscmagic.WrapIRC27NFTMetadata(metadata),
	}
}

// handler for ISCSandbox::getTimestampUnixSeconds
func (h *magicContractViewHandler) GetTimestampUnixSeconds() int64 {
	return h.ctx.Timestamp().Unix()
}

// handler for ISCSandbox::callView
func (h *magicContractViewHandler) CallView(
	contractHname uint32,
	entryPoint uint32,
	params iscmagic.ISCDict,
) iscmagic.ISCDict {
	callRet := h.ctx.CallView(
		isc.Hname(contractHname),
		isc.Hname(entryPoint),
		params.Unwrap(),
	)
	return iscmagic.WrapISCDict(callRet)
}

// handler for ISCSandbox::getAllowanceFrom
func (h *magicContractViewHandler) GetAllowanceFrom(addr common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, addr, h.caller.Address()))
}

// handler for ISCSandbox::getAllowanceTo
func (h *magicContractViewHandler) GetAllowanceTo(target common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, h.caller.Address(), target))
}

// handler for ISCSandbox::getAllowance
func (h *magicContractViewHandler) GetAllowance(from, to common.Address) iscmagic.ISCAssets {
	return iscmagic.WrapISCAssets(getAllowance(h.ctx, from, to))
}

// handler for ISCSandbox::getBaseTokenProperties
func (h *magicContractViewHandler) GetBaseTokenProperties() iscmagic.ISCTokenProperties {
	l1 := parameters.L1()
	return iscmagic.ISCTokenProperties{
		Name:         l1.BaseToken.Name,
		TickerSymbol: l1.BaseToken.TickerSymbol,
		Decimals:     uint8(l1.BaseToken.Decimals),
		TotalSupply:  big.NewInt(int64(l1.Protocol.TokenSupply)),
	}
}

// handler for ISCSandbox::erc20NativeTokensAddress
func (h *magicContractViewHandler) Erc20NativeTokensAddress(foundrySN uint32) common.Address {
	return iscmagic.ERC20NativeTokensAddress(foundrySN)
}

// handler for ISCSandbox::erc20NativeTokensFoundrySerialNumber
func (h *magicContractViewHandler) Erc20NativeTokensFoundrySerialNumber(addr common.Address) uint32 {
	sn, err := iscmagic.ERC20NativeTokensFoundrySN(addr)
	h.ctx.RequireNoError(err)
	return sn
}

// handler for ISCSandbox::getNativeTokenID
func (h *magicContractViewHandler) GetNativeTokenID(foundrySN uint32) iscmagic.NativeTokenID {
	r := h.ctx.CallView(accounts.Contract.Hname(), accounts.ViewFoundryOutput.Hname(), dict.Dict{
		accounts.ParamFoundrySN: codec.EncodeUint32(foundrySN),
	})
	out := &iotago.FoundryOutput{}
	_, err := out.Deserialize(r.MustGet(accounts.ParamFoundryOutputBin), serializer.DeSeriModeNoValidation, nil)
	h.ctx.RequireNoError(err)
	nativeTokenID := out.MustNativeTokenID()
	return iscmagic.WrapNativeTokenID(nativeTokenID)
}

// handler for ISCSandbox::getNativeTokenScheme
func (h *magicContractViewHandler) GetNativeTokenScheme(foundrySN uint32) iotago.SimpleTokenScheme {
	r := h.ctx.CallView(accounts.Contract.Hname(), accounts.ViewFoundryOutput.Hname(), dict.Dict{
		accounts.ParamFoundrySN: codec.EncodeUint32(foundrySN),
	})
	out := &iotago.FoundryOutput{}
	_, err := out.Deserialize(r.MustGet(accounts.ParamFoundryOutputBin), serializer.DeSeriModeNoValidation, nil)
	h.ctx.RequireNoError(err)
	s, ok := out.TokenScheme.(*iotago.SimpleTokenScheme)
	if !ok {
		panic("expected SimpleTokenScheme")
	}
	return *s
}
