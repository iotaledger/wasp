// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"bytes"
	"sort"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

const (
	hasBaseTokens   = 0x80
	hasNativeTokens = 0x40
	hasNFTs         = 0x20
)

type TokenAmounts map[wasmtypes.ScTokenID]wasmtypes.ScBigInt

type ScAssets struct {
	BaseTokens   uint64
	NativeTokens TokenAmounts
	Nfts         map[wasmtypes.ScNftID]bool
}

func NewScAssets(buf []byte) *ScAssets {
	assets := &ScAssets{}
	if len(buf) == 0 {
		return assets
	}

	dec := wasmtypes.NewWasmDecoder(buf)
	flags := wasmtypes.Uint8Decode(dec)
	if flags == 0x00 {
		return assets
	}

	if (flags & hasBaseTokens) != 0 {
		assets.BaseTokens = dec.VluDecode(64)
	}
	if (flags & hasNativeTokens) != 0 {
		size := dec.VluDecode(16)
		assets.NativeTokens = make(TokenAmounts, size)
		for ; size > 0; size-- {
			tokenID := wasmtypes.TokenIDDecode(dec)
			assets.NativeTokens[tokenID] = wasmtypes.BigIntDecode(dec)
		}
	}
	if (flags & hasNFTs) != 0 {
		size := dec.VluDecode(16)
		assets.Nfts = make(map[wasmtypes.ScNftID]bool)
		for ; size > 0; size-- {
			nftID := wasmtypes.NftIDDecode(dec)
			assets.Nfts[nftID] = true
		}
	}
	return assets
}

func (a *ScAssets) Balances() ScBalances {
	return ScBalances{assets: a}
}

func (a *ScAssets) Bytes() []byte {
	if a == nil {
		return []byte{}
	}

	enc := wasmtypes.NewWasmEncoder()
	if a.IsEmpty() {
		return []byte{0}
	}

	var flags byte
	if a.BaseTokens != 0 {
		flags |= hasBaseTokens
	}
	if len(a.NativeTokens) != 0 {
		flags |= hasNativeTokens
	}
	if len(a.Nfts) != 0 {
		flags |= hasNFTs
	}
	wasmtypes.Uint8Encode(enc, flags)

	if (flags & hasBaseTokens) != 0 {
		enc.VluEncode(a.BaseTokens)
	}
	if (flags & hasNativeTokens) != 0 {
		enc.VluEncode(uint64(len(a.NativeTokens)))
		for _, tokenID := range a.TokenIDs() {
			wasmtypes.TokenIDEncode(enc, *tokenID)
			wasmtypes.BigIntEncode(enc, a.NativeTokens[*tokenID])
		}
	}
	if (flags & hasNFTs) != 0 {
		nftIDs := make([]*wasmtypes.ScNftID, 0, len(a.Nfts))
		for key := range a.Nfts {
			// need a local copy to avoid referencing the single key var multiple times
			nftID := key
			nftIDs = append(nftIDs, &nftID)
		}
		sort.Slice(nftIDs, func(i, j int) bool {
			return bytes.Compare(nftIDs[i].Bytes(), nftIDs[j].Bytes()) < 0
		})
		enc.VluEncode(uint64(len(a.Nfts)))
		for _, nftID := range nftIDs {
			wasmtypes.NftIDEncode(enc, *nftID)
		}
	}
	return enc.Buf()
}

func (a *ScAssets) IsEmpty() bool {
	if a.BaseTokens != 0 {
		return false
	}
	for _, val := range a.NativeTokens {
		if !val.IsZero() {
			return false
		}
	}
	return len(a.Nfts) == 0
}

func (a *ScAssets) NftIDs() []*wasmtypes.ScNftID {
	nftIDs := make([]*wasmtypes.ScNftID, 0, len(a.Nfts))
	for key := range a.Nfts {
		// need a local copy to avoid referencing the single key var multiple times
		nftID := key
		nftIDs = append(nftIDs, &nftID)
	}
	sort.Slice(nftIDs, func(i, j int) bool {
		return bytes.Compare(nftIDs[i].Bytes(), nftIDs[j].Bytes()) < 0
	})
	return nftIDs
}

func (a *ScAssets) TokenIDs() []*wasmtypes.ScTokenID {
	tokenIDs := make([]*wasmtypes.ScTokenID, 0, len(a.NativeTokens))
	for key := range a.NativeTokens {
		// need a local copy to avoid referencing the single key var multiple times
		tokenID := key
		tokenIDs = append(tokenIDs, &tokenID)
	}
	sort.Slice(tokenIDs, func(i, j int) bool {
		return bytes.Compare(tokenIDs[i].Bytes(), tokenIDs[j].Bytes()) < 0
	})
	return tokenIDs
}

type ScBalances struct {
	assets *ScAssets
}

func (b *ScBalances) Balance(tokenID *wasmtypes.ScTokenID) wasmtypes.ScBigInt {
	if len(b.assets.NativeTokens) == 0 {
		return wasmtypes.NewScBigInt()
	}
	return b.assets.NativeTokens[*tokenID]
}

func (b *ScBalances) Bytes() []byte {
	if b == nil {
		return []byte{}
	}
	return b.assets.Bytes()
}

func (b *ScBalances) BaseTokens() uint64 {
	return b.assets.BaseTokens
}

func (b *ScBalances) IsEmpty() bool {
	return b.assets.IsEmpty()
}

func (b *ScBalances) NftIDs() []*wasmtypes.ScNftID {
	return b.assets.NftIDs()
}

func (b *ScBalances) TokenIDs() []*wasmtypes.ScTokenID {
	return b.assets.TokenIDs()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransfer struct {
	ScBalances
}

// create a new transfer object ready to add token transfers
func NewScTransfer() *ScTransfer {
	return &ScTransfer{ScBalances{assets: NewScAssets(nil)}}
}

// create a new transfer object from a balances object
func ScTransferFromBalances(balances *ScBalances) *ScTransfer {
	transfer := ScTransferFromBaseTokens(balances.BaseTokens())
	for _, tokenID := range balances.TokenIDs() {
		transfer.Set(tokenID, balances.Balance(tokenID))
	}
	nftIDs := balances.NftIDs()
	for i := range nftIDs {
		transfer.AddNFT(nftIDs[i])
	}
	return transfer
}

// create a new transfer object and initialize it with the specified amount of base tokens
func ScTransferFromBaseTokens(amount uint64) *ScTransfer {
	transfer := NewScTransfer()
	transfer.assets.BaseTokens = amount
	return transfer
}

// create a new transfer object and initialize it with the specified NFT
func ScTransferFromNFT(nftID *wasmtypes.ScNftID) *ScTransfer {
	transfer := NewScTransfer()
	transfer.AddNFT(nftID)
	return transfer
}

// create a new transfer object and initialize it with the specified token transfer
func ScTransferFromTokens(tokenID *wasmtypes.ScTokenID, amount wasmtypes.ScBigInt) *ScTransfer {
	transfer := NewScTransfer()
	transfer.Set(tokenID, amount)
	return transfer
}

func (t *ScTransfer) AddNFT(nftID *wasmtypes.ScNftID) {
	if t.assets.Nfts == nil {
		t.assets.Nfts = make(map[wasmtypes.ScNftID]bool)
	}
	t.assets.Nfts[*nftID] = true
}

func (t *ScTransfer) Bytes() []byte {
	if t == nil {
		return []byte{}
	}
	return t.assets.Bytes()
}

// set the specified tokenID amount in the transfers object
// note that this will overwrite any previous amount for the specified tokenID
func (t *ScTransfer) Set(tokenID *wasmtypes.ScTokenID, amount wasmtypes.ScBigInt) {
	if t.assets.NativeTokens == nil {
		t.assets.NativeTokens = make(TokenAmounts)
	}
	t.assets.NativeTokens[*tokenID] = amount
}
