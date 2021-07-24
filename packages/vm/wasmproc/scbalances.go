// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/iscp/color"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func NewScBalances(vm *WasmProcessor, keyID int32) *ScDict {
	o := NewScDict(&vm.KvStoreHost, dict.New())
	switch keyID {
	case wasmhost.KeyIncoming:
		if vm.ctx == nil {
			o.Panic("no incoming() on views")
		}
		return loadBalances(o, vm.ctx.IncomingTransfer())
	case wasmhost.KeyMinted:
		if vm.ctx == nil {
			o.Panic("no minted() on views")
		}
		return loadBalances(o, vm.ctx.Minted())

	case wasmhost.KeyBalances:
		if vm.ctx != nil {
			return loadBalances(o, vm.ctx.Balances())
		}
		return loadBalances(o, vm.ctxView.Balances())
	}
	o.Panic("unknown balances: %s", vm.GetKeyStringFromID(keyID))
	return nil
}

func loadBalances(o *ScDict, balances color.Balances) *ScDict {
	index := 0
	key := o.host.GetKeyStringFromID(wasmhost.KeyColor)
	balances.ForEachRandomly(func(color color.Color, balance uint64) bool {
		o.kvStore.Set(kv.Key(color[:]), codec.EncodeUint64(balance))
		o.kvStore.Set(kv.Key(key+"."+strconv.Itoa(index)), color[:])
		index++
		return true
	})
	// save KeyLength
	o.kvStore.Set(kv.Key(key), codec.EncodeInt32(int32(index)))
	return o
}
