// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type ScContext struct {
	ScDict
}

func NewScContext(vm *wasmProcessor) *ScContext {
	o := &ScContext{}
	o.name = "root"
	o.id = 1
	o.isRoot = true
	o.vm = vm
	o.objects = make(map[int32]int32)
	return o
}

func (o *ScContext) Exists(keyId int32) bool {
	if keyId == wasmhost.KeyExports {
		return o.vm.ctx == nil && o.vm.ctxView == nil
	}
	return o.GetTypeId(keyId) > 0
}

func (o *ScContext) Finalize() {
	o.objects = make(map[int32]int32)
	o.vm.ResetObjects()
}

func (o *ScContext) GetBytes(keyId int32) []byte {
	switch keyId {
	case wasmhost.KeyCaller:
		id := o.vm.ctx.Caller()
		return id.Bytes()
	}
	return o.ScDict.GetBytes(keyId)
}

func (o *ScContext) GetInt(keyId int32) int64 {
	switch keyId {
	case wasmhost.KeyTimestamp:
		return o.vm.ctx.GetTimestamp()
	}
	return o.ScDict.GetInt(keyId)
}

func (o *ScContext) GetObjectId(keyId int32, typeId int32) int32 {
	if keyId == wasmhost.KeyExports && (o.vm.ctx != nil || o.vm.ctxView != nil) {
		// once map has entries (after on_load) this cannot be called any more
		return o.ScDict.GetObjectId(keyId, typeId)
	}

	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		wasmhost.KeyBalances:  func() WaspObject { return &ScBalances{} },
		wasmhost.KeyCalls:     func() WaspObject { return &ScCalls{} },
		wasmhost.KeyContract:  func() WaspObject { return &ScContract{} },
		wasmhost.KeyExports:   func() WaspObject { return &ScExports{} },
		wasmhost.KeyIncoming:  func() WaspObject { return &ScBalances{incoming: true} },
		wasmhost.KeyLogs:      func() WaspObject { return &ScLogs{} },
		wasmhost.KeyParams:    func() WaspObject { return &ScDict{kvStore: o.vm.Params()} },
		wasmhost.KeyPosts:     func() WaspObject { return &ScPosts{} },
		wasmhost.KeyResults:   func() WaspObject { return &ScDict{kvStore: dict.New()} },
		wasmhost.KeyState:     func() WaspObject { return &ScDict{kvStore: o.vm.State()} },
		wasmhost.KeyTransfers: func() WaspObject { return &ScTransfers{} },
		wasmhost.KeyUtility:   func() WaspObject { return &ScUtility{} },
		wasmhost.KeyViews:     func() WaspObject { return &ScViews{} },
	})
}

func (o *ScContext) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyBalances:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyCalls:
		return wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY
	case wasmhost.KeyContract:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyExports:
		return wasmhost.OBJTYPE_STRING | wasmhost.OBJTYPE_ARRAY
	case wasmhost.KeyIncoming:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyLogs:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyParams:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyPosts:
		return wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY
	case wasmhost.KeyResults:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyCaller:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_AGENT
	case wasmhost.KeyState:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyTimestamp:
		return wasmhost.OBJTYPE_INT
	case wasmhost.KeyTransfers:
		return wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY
	case wasmhost.KeyUtility:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyViews:
		return wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY
	}
	return 0
}

func (o *ScContext) SetString(keyId int32, value string) {
	switch keyId {
	case wasmhost.KeyEvent:
		o.vm.ctx.Event(value)
	case wasmhost.KeyLog:
		o.vm.log().Infof(value)
	case wasmhost.KeyTrace:
		o.vm.log().Infof(value)
	case wasmhost.KeyPanic:
		o.vm.log().Infof(value)
	}
}
