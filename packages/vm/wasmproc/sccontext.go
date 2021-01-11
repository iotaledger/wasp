// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type ScContext struct {
	ScSandboxObject
}

func NewScContext(vm *wasmProcessor) *ScContext {
	o := &ScContext{}
	o.vm = vm
	o.host = &vm.KvStoreHost
	o.name = "root"
	o.id = 1
	o.isRoot = true
	o.objects = make(map[int32]int32)
	return o
}

func (o *ScContext) Exists(keyId int32) bool {
	if keyId == wasmhost.KeyExports {
		return o.vm.ctx == nil && o.vm.ctxView == nil
	}
	return o.GetTypeId(keyId) > 0
}

func (o *ScContext) finalize() {
	o.objects = make(map[int32]int32)
	o.host.ResetObjects()
}

func (o *ScContext) GetBytes(keyId int32) []byte {
	switch keyId {
	case wasmhost.KeyCaller:
		id := o.vm.ctx.Caller()
		return id[:]
	}
	o.invalidKey(keyId)
	return nil
}

func (o *ScContext) GetInt(keyId int32) int64 {
	switch keyId {
	case wasmhost.KeyTimestamp:
		return o.vm.ctx.GetTimestamp()
	}
	o.invalidKey(keyId)
	return 0
}

func (o *ScContext) GetObjectId(keyId int32, typeId int32) int32 {
	if keyId == wasmhost.KeyExports && (o.vm.ctx != nil || o.vm.ctxView != nil) {
		// once map has entries (after on_load) this cannot be called any more
		o.invalidKey(keyId)
		return 0
	}

	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		wasmhost.KeyBalances:  func() WaspObject { return NewScBalances(o.vm, false) },
		wasmhost.KeyCalls:     func() WaspObject { return NewScCalls(o.vm) },
		wasmhost.KeyContract:  func() WaspObject { return NewScContract(o.vm) },
		wasmhost.KeyDeploys:   func() WaspObject { return NewScDeploys(o.vm) },
		wasmhost.KeyExports:   func() WaspObject { return NewScExports(o.vm) },
		wasmhost.KeyIncoming:  func() WaspObject { return NewScBalances(o.vm, true) },
		wasmhost.KeyLogs:      func() WaspObject { return NewScLogs(o.vm) },
		wasmhost.KeyParams:    func() WaspObject { return NewScDict(o.vm, o.vm.params()) },
		wasmhost.KeyPosts:     func() WaspObject { return NewScPosts(o.vm) },
		wasmhost.KeyResults:   func() WaspObject { return NewScDict(o.vm, nil) },
		wasmhost.KeyState:     func() WaspObject { return NewScDict(o.vm, o.vm.state()) },
		wasmhost.KeyTransfers: func() WaspObject { return NewScTransfers(o.vm) },
		wasmhost.KeyUtility:   func() WaspObject { return NewScUtility(o.vm) },
		wasmhost.KeyViews:     func() WaspObject { return NewScViews(o.vm) },
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
		o.vm.log().Debugf(value)
	case wasmhost.KeyPanic:
		o.vm.log().Panicf(value)
	default:
		o.invalidKey(keyId)
	}
}
