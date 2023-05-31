// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCAccounts::getL2BalanceBaseTokens
func (h *magicContractHandler) GetL2BalanceBaseTokens(agentID iscmagic.ISCAgentID) uint64 {
	r := h.callView(accounts.Contract.Hname(), accounts.ViewBalanceBaseToken.Hname(), dict.Dict{
		accounts.ParamAgentID: codec.EncodeAgentID(agentID.MustUnwrap()),
	})
	return codec.MustDecodeUint64(r.Get(accounts.ParamBalance))
}

// handler for ISCAccounts::getL2BalanceNativeTokens
func (h *magicContractHandler) GetL2BalanceNativeTokens(nativeTokenID iscmagic.NativeTokenID, agentID iscmagic.ISCAgentID) *big.Int {
	r := h.callView(accounts.Contract.Hname(), accounts.ViewBalanceNativeToken.Hname(), dict.Dict{
		accounts.ParamNativeTokenID: codec.EncodeNativeTokenID(nativeTokenID.Unwrap()),
		accounts.ParamAgentID:       codec.EncodeAgentID(agentID.MustUnwrap()),
	})
	return codec.MustDecodeBigIntAbs(r.Get(accounts.ParamBalance))
}

// handler for ISCAccounts::getL2NFTs
func (h *magicContractHandler) GetL2NFTs(agentID iscmagic.ISCAgentID) []iscmagic.NFTID {
	r := h.callView(
		accounts.Contract.Hname(),
		accounts.ViewAccountNFTs.Hname(),
		dict.Dict{accounts.ParamAgentID: codec.EncodeAgentID(agentID.MustUnwrap())},
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
		dict.Dict{accounts.ParamAgentID: codec.EncodeAgentID(agentID.MustUnwrap())},
	)
	n := codec.MustDecodeUint32(r[accounts.ParamNFTAmount])
	return big.NewInt(int64(n))
}

// handler for ISCAccounts::getL2NFTsInCollection
func (h *magicContractHandler) GetL2NFTsInCollection(agentID iscmagic.ISCAgentID, collectionID iscmagic.NFTID) []iscmagic.NFTID {
	r := h.callView(
		accounts.Contract.Hname(),
		accounts.ViewAccountNFTsInCollection.Hname(),
		dict.Dict{
			accounts.ParamAgentID:      codec.EncodeAgentID(agentID.MustUnwrap()),
			accounts.ParamCollectionID: codec.EncodeNFTID(collectionID.Unwrap()),
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
			accounts.ParamAgentID:      codec.EncodeAgentID(agentID.MustUnwrap()),
			accounts.ParamCollectionID: codec.EncodeNFTID(collectionID.Unwrap()),
		},
	)
	n := codec.MustDecodeUint32(r[accounts.ParamNFTAmount])
	return big.NewInt(int64(n))
}

// handler for ISCAccounts::foundryCreateNew
func (h *magicContractHandler) FoundryCreateNew(tokenScheme iotago.SimpleTokenScheme, allowance iscmagic.ISCAssets) uint32 {
	ret := h.call(
		accounts.Contract.Hname(),
		accounts.FuncFoundryCreateNew.Hname(),
		dict.Dict{
			accounts.ParamTokenScheme: codec.EncodeTokenScheme(&tokenScheme),
		},
		allowance.Unwrap(),
	)
	return codec.MustDecodeUint32(ret.Get(accounts.ParamFoundrySN))
}

// handler for ISCAccounts::mintBaseTokens
func (h *magicContractHandler) MintNativeTokens(foundrySN uint32, amount *big.Int, allowance iscmagic.ISCAssets) {
	h.call(
		accounts.Contract.Hname(),
		accounts.FuncFoundryModifySupply.Hname(),
		dict.Dict{
			accounts.ParamFoundrySN:      codec.EncodeUint32(foundrySN),
			accounts.ParamSupplyDeltaAbs: codec.EncodeBigIntAbs(amount),
			accounts.ParamDestroyTokens:  codec.EncodeBool(false),
		},
		allowance.Unwrap(),
	)
}
