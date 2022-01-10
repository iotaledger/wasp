// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func NewScBalances(wc *WasmContext, keyID int32) *ScDict {
	panic("TODO implement")
	// o := NewScDict(&wc.KvStoreHost, dict.New())
	// switch keyID {
	// case wasmhost.KeyIncoming:
	// 	if wc.ctx == nil {
	// 		o.Panicf("no incoming() on views")
	// 	}
	// 	return loadBalances(o, wc.ctx.Allowance())
	// case wasmhost.KeyMinted:
	// 	if wc.ctx == nil {
	// 		o.Panicf("no minted() on views")
	// 	}
	// 	return loadBalances(o, wc.ctx.Minted())

	// case wasmhost.KeyBalances:
	// 	if wc.ctx != nil {
	// 		return loadBalances(o, wc.ctx.Balances())
	// 	}
	// 	return loadBalances(o, wc.ctxView.Balances())
	// }
	// o.Panicf("unknown balances: %s", wc.GetKeyStringFromID(keyID))
	// return nil
}

func loadBalances(o *ScDict, balances *iscp.Assets) *ScDict {
	panic("TODO implement")
	// index := 0
	// key := o.host.GetKeyStringFromID(wasmhost.KeyColor)
	// balances.ForEachRandomly(func(assetID []byte, balance uint64) bool {
	// 	o.kvStore.Set(kv.Key(assetID[:]), codec.EncodeUint64(balance))
	// 	o.kvStore.Set(kv.Key(key+"."+strconv.Itoa(index)), assetID[:])
	// 	index++
	// 	return true
	// })
	// // save KeyLength
	// o.kvStore.Set(kv.Key(key), codec.EncodeInt32(int32(index)))
	// return o
}
