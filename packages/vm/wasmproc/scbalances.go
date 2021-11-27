// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func NewScBalances(wc *WasmContext, keyID int32) *ScDict {
	o := NewScDict(&wc.KvStoreHost, dict.New())
	switch keyID {
	case wasmhost.KeyIncoming:
		if wc.ctx == nil {
			o.Panicf("no incoming() on views")
		}
		return loadBalances(o, wc.ctx.IncomingTransfer())
	case wasmhost.KeyMinted:
		if wc.ctx == nil {
			o.Panicf("no minted() on views")
		}
		return loadBalances(o, wc.ctx.Minted())

	case wasmhost.KeyBalances:
		if wc.ctx != nil {
			return loadBalances(o, wc.ctx.Balances())
		}
		return loadBalances(o, wc.ctxView.Balances())
	}
	o.Panicf("unknown balances: %s", wc.GetKeyStringFromID(keyID))
	return nil
}

func loadBalances(o *ScDict, balances colored.Balances) *ScDict {
	index := 0
	key := o.host.GetKeyStringFromID(wasmhost.KeyColor)
	balances.ForEachRandomly(func(color colored.Color, balance uint64) bool {
		o.kvStore.Set(kv.Key(color[:]), codec.EncodeUint64(balance))
		o.kvStore.Set(kv.Key(key+"."+strconv.Itoa(index)), color[:])
		index++
		return true
	})
	// save KeyLength
	o.kvStore.Set(kv.Key(key), codec.EncodeInt32(int32(index)))
	return o
}
