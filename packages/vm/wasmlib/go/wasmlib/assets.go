// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"sort"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

type ScAssets map[wasmtypes.ScColor]uint64

func NewScAssetsFromBytes(buf []byte) ScAssets {
	dict := make(ScAssets)
	size, buf := wasmcodec.ExtractUint32(buf)
	var k []byte
	var v uint64
	for i := uint32(0); i < size; i++ {
		k, buf = wasmcodec.ExtractBytes(buf, 32)
		v, buf = wasmcodec.ExtractUint64(buf)
		dict[wasmtypes.ColorFromBytes(k)] = v
	}
	return dict
}

func (a ScAssets) Bytes() []byte {
	keys := make([]wasmtypes.ScColor, 0, len(a))
	for key := range a {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i].Bytes()) < string(keys[j].Bytes())
	})
	buf := wasmcodec.AppendUint32(nil, uint32(len(keys)))
	for _, k := range keys {
		v := a[k]
		buf = wasmcodec.AppendBytes(buf, k.Bytes())
		buf = wasmcodec.AppendUint64(buf, v)
	}
	return buf
}

func (a ScAssets) Balances() ScBalances {
	return ScBalances{assets: a}
}

type ScBalances struct {
	assets ScAssets
}

func (b ScBalances) Balance(color wasmtypes.ScColor) uint64 {
	return b.assets[color]
}

func (b ScBalances) Colors() []wasmtypes.ScColor {
	res := make([]wasmtypes.ScColor, 0, len(b.assets))
	for color := range b.assets {
		res = append(res, color)
	}
	return res
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransfers ScAssets

// create a new transfers object ready to add token transfers
func NewScTransfers() ScTransfers {
	return make(ScTransfers)
}

// create a new transfers object from a balances object
func NewScTransfersFromBalances(balances ScBalances) ScTransfers {
	return ScTransfers(balances.assets)
}

// create a new transfers object and initialize it with the specified amount of iotas
func NewScTransferIotas(amount uint64) ScTransfers {
	return NewScTransfer(wasmtypes.IOTA, amount)
}

// create a new transfers object and initialize it with the specified token transfer
func NewScTransfer(color wasmtypes.ScColor, amount uint64) ScTransfers {
	transfer := make(ScTransfers)
	transfer[color] = amount
	return transfer
}

// set the specified colored token transfer in the transfers object
// note that this will overwrite any previous amount for the specified color
func (t ScTransfers) Set(color wasmtypes.ScColor, amount uint64) {
	t[color] = amount
}
