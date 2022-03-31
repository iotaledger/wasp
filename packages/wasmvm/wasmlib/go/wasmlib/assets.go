// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"sort"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type TokenAmounts map[wasmtypes.ScTokenID]wasmtypes.ScBigInt

type ScAssets struct {
	Iotas  uint64
	Tokens TokenAmounts
	NFTs   []*wasmtypes.ScNftID
}

func NewScAssetsFromBytes(buf []byte) *ScAssets {
	assets := &ScAssets{}
	if len(buf) == 0 {
		return assets
	}
	dec := wasmtypes.NewWasmDecoder(buf)
	assets.Iotas = wasmtypes.Uint64Decode(dec)

	size := wasmtypes.Uint32Decode(dec)
	if size > 0 {
		assets.Tokens = make(TokenAmounts, size)
		for ; size > 0; size-- {
			tokenID := wasmtypes.TokenIDDecode(dec)
			assets.Tokens[tokenID] = wasmtypes.BigIntDecode(dec)
		}
	}

	size = wasmtypes.Uint32Decode(dec)
	for ; size > 0; size-- {
		nftID := wasmtypes.NftIDDecode(dec)
		assets.NFTs = append(assets.NFTs, &nftID)
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
	wasmtypes.Uint64Encode(enc, a.Iotas)

	wasmtypes.Uint32Encode(enc, uint32(len(a.Tokens)))
	for _, tokenID := range a.TokenIDs() {
		wasmtypes.TokenIDEncode(enc, *tokenID)
		wasmtypes.BigIntEncode(enc, a.Tokens[*tokenID])
	}

	wasmtypes.Uint32Encode(enc, uint32(len(a.NFTs)))
	for _, nftID := range a.NFTs {
		wasmtypes.NftIDEncode(enc, *nftID)
	}
	return enc.Buf()
}

func (a *ScAssets) IsEmpty() bool {
	if a.Iotas != 0 {
		return false
	}
	for _, val := range a.Tokens {
		if !val.IsZero() {
			return false
		}
	}
	return len(a.NFTs) == 0
}

func (a *ScAssets) TokenIDs() []*wasmtypes.ScTokenID {
	keys := make([]*wasmtypes.ScTokenID, 0, len(a.Tokens))
	for key := range a.Tokens {
		keys = append(keys, &key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i].Bytes()) < string(keys[j].Bytes())
	})
	return keys
}

type ScBalances struct {
	assets *ScAssets
}

func (b *ScBalances) Balance(tokenID *wasmtypes.ScTokenID) wasmtypes.ScBigInt {
	return b.assets.Tokens[*tokenID]
}

func (b *ScBalances) Bytes() []byte {
	if b == nil {
		return []byte{}
	}
	return b.assets.Bytes()
}

func (b *ScBalances) Iotas() uint64 {
	return b.assets.Iotas
}

func (b *ScBalances) IsEmpty() bool {
	return b.assets.IsEmpty()
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
	return &ScTransfer{ScBalances{assets: &ScAssets{}}}
}

// create a new transfer object from a balances object
func NewScTransferFromBalances(balances *ScBalances) *ScTransfer {
	transfer := NewScTransfer()
	tokenIDs := balances.TokenIDs()
	for _, tokenID := range tokenIDs {
		transfer.Set(tokenID, balances.Balance(tokenID))
	}
	return transfer
}

// create a new transfer object and initialize it with the specified amount of iotas
func NewScTransferIotas(amount uint64) *ScTransfer {
	transfer := NewScTransfer()
	transfer.assets.Iotas = amount
	return transfer
}

// create a new transfer object and initialize it with the specified token transfer
func NewScTransferToken(tokenID *wasmtypes.ScTokenID, amount wasmtypes.ScBigInt) *ScTransfer {
	transfer := NewScTransfer()
	transfer.Set(tokenID, amount)
	return transfer
}

// create a new transfer object and initialize it with the specified NFT
func NewScTransferNFT(nftID *wasmtypes.ScNftID) *ScTransfer {
	transfer := NewScTransfer()
	transfer.assets.NFTs = append(transfer.assets.NFTs, nftID)
	return transfer
}

func (t *ScTransfer) AddNFT(nftID *wasmtypes.ScNftID) {
	// TODO filter doubles
	t.assets.NFTs = append(t.assets.NFTs, nftID)
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
	t.assets.Tokens[*tokenID] = amount
}
