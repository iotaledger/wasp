// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"strconv"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func NewScBalances(vm *wasmProcessor, incoming bool) *ScDict {
	o := NewScDict(vm)
	if incoming {
		if vm.ctx == nil {
			o.Panic("No incoming transfers on views")
		}
		return loadBalances(o, vm.ctx.IncomingTransfer())
	}
	if vm.ctx != nil {
		return loadBalances(o, vm.ctx.Balances())
	}
	return loadBalances(o, vm.ctxView.Balances())
}

func loadBalances(o *ScDict, balances coretypes.ColoredBalances) *ScDict {
	index := 0
	balances.IterateDeterministic(func(color balance.Color, balance int64) bool {
		o.kvStore.Set(kv.Key(color[:]), codec.EncodeInt64(balance))
		o.kvStore.Set(kv.Key("color." + strconv.Itoa(index)), color[:])
		index++
		return true
	})
	o.kvStore.Set("color", codec.EncodeInt64(int64(index)))
	return o
}
