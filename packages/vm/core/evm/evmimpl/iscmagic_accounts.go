// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCAccounts::getL2BalanceBaseTokens
func (h *magicContractHandler) GetL2BalanceBaseTokens(agentID iscmagic.ISCAgentID) uint64 {
	r := h.callView(accounts.Contract.Hname(), accounts.ViewBalanceBaseToken.Hname(), dict.Dict{
		accounts.ParamAgentID: codec.AgentID.Encode(agentID.MustUnwrap()),
	})
	return codec.Uint64.MustDecode(r.Get(accounts.ParamBalance))
}

// handler for ISCAccounts::getL2BalanceNativeTokens
func (h *magicContractHandler) GetL2BalanceNativeTokens(nativeTokenID iscmagic.NativeTokenID, agentID iscmagic.ISCAgentID) *big.Int {
	r := h.callView(accounts.Contract.Hname(), accounts.ViewBalanceNativeToken.Hname(), dict.Dict{
		accounts.ParamNativeTokenID: codec.NativeTokenID.Encode(nativeTokenID.Unwrap()),
		accounts.ParamAgentID:       codec.AgentID.Encode(agentID.MustUnwrap()),
	})
	return codec.BigIntAbs.MustDecode(r.Get(accounts.ParamBalance))
}

// handler for ISCAccounts::getL2NFTs
func (h *magicContractHandler) GetL2NFTs(agentID iscmagic.ISCAgentID) []iscmagic.NFTID {
	r := h.callView(
		accounts.Contract.Hname(),
		accounts.ViewAccountNFTs.Hname(),
		dict.Dict{accounts.ParamAgentID: codec.AgentID.Encode(agentID.MustUnwrap())},
	)
	nftIDs := collections.NewArray(r, accounts.ParamNFTIDs)
	ret := make([]iscmagic.NFTID, nftIDs.Len())
	for i := range ret {
		copy(ret[i][:], nftIDs.GetAt(uint32(i)))
	}
	return ret
}

// handler for ISCAccounts::getL2NFTAmount
func (h *magicContractHandler) GetL2NFTAmount(agentID iscmagic.ISCAgentID) *big.Int {
	r := h.callView(
		accounts.Contract.Hname(),
		accounts.ViewAccountNFTAmount.Hname(),
		dict.Dict{accounts.ParamAgentID: codec.AgentID.Encode(agentID.MustUnwrap())},
	)
	n := codec.Uint32.MustDecode(r[accounts.ParamNFTAmount])
	return big.NewInt(int64(n))
}

// handler for ISCAccounts::getL2NFTsInCollection
func (h *magicContractHandler) GetL2NFTsInCollection(agentID iscmagic.ISCAgentID, collectionID iscmagic.NFTID) []iscmagic.NFTID {
	r := h.callView(
		accounts.Contract.Hname(),
		accounts.ViewAccountNFTsInCollection.Hname(),
		dict.Dict{
			accounts.ParamAgentID:      codec.AgentID.Encode(agentID.MustUnwrap()),
			accounts.ParamCollectionID: codec.NFTID.Encode(collectionID.Unwrap()),
		},
	)
	nftIDs := collections.NewArray(r, accounts.ParamNFTIDs)
	ret := make([]iscmagic.NFTID, nftIDs.Len())
	for i := range ret {
		copy(ret[i][:], nftIDs.GetAt(uint32(i)))
	}
	return ret
}

// handler for ISCAccounts::getL2NFTAmountInCollection
func (h *magicContractHandler) GetL2NFTAmountInCollection(agentID iscmagic.ISCAgentID, collectionID iscmagic.NFTID) *big.Int {
	r := h.callView(
		accounts.Contract.Hname(),
		accounts.ViewAccountNFTAmountInCollection.Hname(),
		dict.Dict{
			accounts.ParamAgentID:      codec.AgentID.Encode(agentID.MustUnwrap()),
			accounts.ParamCollectionID: codec.NFTID.Encode(collectionID.Unwrap()),
		},
	)
	n := codec.Uint32.MustDecode(r[accounts.ParamNFTAmount])
	return big.NewInt(int64(n))
}

// handler for ISCAccounts::foundryCreateNew
func (h *magicContractHandler) FoundryCreateNew(tokenScheme iotago.SimpleTokenScheme, allowance iscmagic.ISCAssets) uint32 {
	ret := h.call(
		accounts.Contract.Hname(),
		accounts.FuncFoundryCreateNew.Hname(),
		dict.Dict{
			accounts.ParamTokenScheme: codec.TokenScheme.Encode(&tokenScheme),
		},
		allowance.Unwrap(),
	)
	return codec.Uint32.MustDecode(ret.Get(accounts.ParamFoundrySN))
}

func (h *magicContractHandler) CreateNativeTokenFoundry(tokenName string, tickerSymbol string, decimals uint8, tokenScheme iotago.SimpleTokenScheme, allowance iscmagic.ISCAssets) uint32 {
	ret := h.ctx.Privileged().CallOnBehalfOf(
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), h.caller.Address()),
		accounts.Contract.Hname(),
		accounts.FuncNativeTokenCreate.Hname(),
		dict.Dict{
			accounts.ParamTokenScheme:       codec.TokenScheme.Encode(&tokenScheme),
			accounts.ParamTokenName:         codec.String.Encode(tokenName),
			accounts.ParamTokenTickerSymbol: codec.String.Encode(tickerSymbol),
			accounts.ParamTokenDecimals:     codec.Uint8.Encode(decimals),
		},
		allowance.Unwrap(),
	)
	return codec.Uint32.MustDecode(ret.Get(accounts.ParamFoundrySN))
}

// handler for ISCAccounts::mintBaseTokens
func (h *magicContractHandler) MintNativeTokens(foundrySN uint32, amount *big.Int, allowance iscmagic.ISCAssets) {
	h.call(
		accounts.Contract.Hname(),
		accounts.FuncNativeTokenModifySupply.Hname(),
		dict.Dict{
			accounts.ParamFoundrySN:      codec.Uint32.Encode(foundrySN),
			accounts.ParamSupplyDeltaAbs: codec.BigIntAbs.Encode(amount),
			accounts.ParamDestroyTokens:  codec.Bool.Encode(false),
		},
		allowance.Unwrap(),
	)
}
