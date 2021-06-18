// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"strconv"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func NewScBalances(vm *wasmProcessor, keyId int32) *ScDict {
	o := NewScDict(vm)
	switch keyId {
	case wasmhost.KeyIncoming:
		if vm.ctx == nil {
			o.Panic("no incoming() on views")
		}
		return loadBalances(o, vm.ctx.IncomingTransfer())
	case wasmhost.KeyMinted:
		if vm.ctx == nil {
			o.Panic("no minted() on views")
		}
		return loadBalances(o, ledgerstate.NewColoredBalances(vm.ctx.Minted()))

	case wasmhost.KeyBalances:
		if vm.ctx != nil {
			return loadBalances(o, vm.ctx.Balances())
		}
		return loadBalances(o, vm.ctxView.Balances())
	}
	o.Panic("unknown balances: %s", vm.GetKeyStringFromId(keyId))
	return nil
}

func loadBalances(o *ScDict, balances *ledgerstate.ColoredBalances) *ScDict {
	index := 0
	key := o.host.GetKeyStringFromId(wasmhost.KeyColor)
	balances.ForEach(func(color ledgerstate.Color, balance uint64) bool {
		o.kvStore.Set(kv.Key(color[:]), codec.EncodeUint64(balance))
		o.kvStore.Set(kv.Key(key+"."+strconv.Itoa(index)), color[:])
		index++
		return true
	})
	o.kvStore.Set(kv.Key(key), codec.EncodeInt64(int64(index)))
	return o
}
