// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScBalances struct {
	ScSandboxObject
	incoming bool
}

func NewScBalances(vm *wasmProcessor, incoming bool) *ScBalances {
	o := &ScBalances{}
	o.vm = vm
	o.incoming = incoming
	return o
}

func (o *ScBalances) Exists(keyId int32) bool {
	return o.GetInt(keyId) != 0
}

func (o *ScBalances) GetBytes(keyId int32) []byte {
	key := o.host.GetKeyFromId(keyId)
	color, _, err := balance.ColorFromBytes(key)
	if err != nil {
		o.Panic("GetBytes: %v", err)
	}
	if color != balance.ColorNew {
		o.Panic("GetBytes: Expected ColorNew")
	}
	id := o.vm.ctx.RequestID()
	return id[:32]
}

func (o *ScBalances) GetInt(keyId int32) int64 {
	key := o.host.GetKeyFromId(keyId)
	color, _, err := balance.ColorFromBytes(key)
	if err != nil {
		o.Panic("GetInt: %v", err)
	}
	balances := o.vm.balances()
	if o.incoming {
		if o.vm.ctx == nil {
			return 0
		}
		balances = o.vm.ctx.IncomingTransfer()
	}
	return balances.Balance(color)
}

func (o *ScBalances) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		wasmhost.KeyColor: func() WaspObject { return NewScBalanceColors(o.vm, o.incoming) },
	})
}

func (o *ScBalances) GetTypeId(keyId int32) int32 {
	if o.Exists(keyId) {
		return wasmhost.OBJTYPE_INT
	}
	return 0
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScBalanceColors struct {
	ScSandboxObject
	colors   []balance.Color
	incoming bool
}

func NewScBalanceColors(vm *wasmProcessor, incoming bool) *ScBalanceColors {
	o := &ScBalanceColors{}
	o.vm = vm
	o.incoming = incoming
	return o
}

func (o *ScBalanceColors) Exists(keyId int32) bool {
	o.loadColors()
	return keyId >= 0 && keyId < int32(len(o.colors))
}

func (o *ScBalanceColors) GetBytes(keyId int32) []byte {
	if !o.Exists(keyId) {
		o.invalidKey(keyId)
	}
	return o.colors[keyId][:]
}

func (o *ScBalanceColors) GetTypeId(keyId int32) int32 {
	if o.Exists(keyId) {
		return wasmhost.OBJTYPE_COLOR
	}
	return 0
}

func (o *ScBalanceColors) loadColors() {
	if len(o.colors) > 0 {
		return
	}
	balances := o.vm.balances()
	if o.incoming {
		if o.vm.ctx == nil {
			return
		}
		balances = o.vm.ctx.IncomingTransfer()
	}
	balances.IterateDeterministic(func(color balance.Color, balance int64) bool {
		o.colors = append(o.colors, color)
		return true
	})
}
