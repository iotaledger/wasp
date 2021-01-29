// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

var typeIds = map[int32]int32{
	wasmhost.KeyBalances:   wasmhost.OBJTYPE_MAP,
	wasmhost.KeyCaller:     wasmhost.OBJTYPE_AGENT,
	wasmhost.KeyCalls:      wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyChainOwner: wasmhost.OBJTYPE_AGENT,
	wasmhost.KeyCreator:    wasmhost.OBJTYPE_AGENT,
	wasmhost.KeyDeploys:    wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyEvent:      wasmhost.OBJTYPE_STRING,
	wasmhost.KeyExports:    wasmhost.OBJTYPE_STRING | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyId:         wasmhost.OBJTYPE_CONTRACT,
	wasmhost.KeyIncoming:   wasmhost.OBJTYPE_MAP,
	wasmhost.KeyLog:        wasmhost.OBJTYPE_STRING,
	wasmhost.KeyLogs:       wasmhost.OBJTYPE_MAP,
	wasmhost.KeyMaps:       wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyPanic:      wasmhost.OBJTYPE_STRING,
	wasmhost.KeyParams:     wasmhost.OBJTYPE_MAP,
	wasmhost.KeyResults:    wasmhost.OBJTYPE_MAP,
	wasmhost.KeyState:      wasmhost.OBJTYPE_MAP,
	wasmhost.KeyTimestamp:  wasmhost.OBJTYPE_INT,
	wasmhost.KeyTrace:      wasmhost.OBJTYPE_STRING,
	wasmhost.KeyTransfers:  wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyUtility:    wasmhost.OBJTYPE_MAP,
}

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

func (o *ScContext) Exists(keyId int32, typeId int32) bool {
	if keyId == wasmhost.KeyExports {
		return o.vm.ctx == nil && o.vm.ctxView == nil
	}
	return o.GetTypeId(keyId) > 0
}

func (o *ScContext) GetBytes(keyId int32, typeId int32) []byte {
	switch keyId {
	case wasmhost.KeyCaller:
		return o.vm.ctx.Caller().Bytes()
	case wasmhost.KeyChainOwner:
		return o.vm.chainOwnerID().Bytes()
	case wasmhost.KeyCreator:
		return o.vm.contractCreator().Bytes()
	case wasmhost.KeyId:
		return o.vm.contractID().Bytes()
	case wasmhost.KeyTimestamp:
		return codec.EncodeInt64(o.vm.ctx.GetTimestamp())
	}
	o.invalidKey(keyId)
	return nil
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
		wasmhost.KeyDeploys:   func() WaspObject { return NewScDeploys(o.vm) },
		wasmhost.KeyExports:   func() WaspObject { return NewScExports(o.vm) },
		wasmhost.KeyIncoming:  func() WaspObject { return NewScBalances(o.vm, true) },
		wasmhost.KeyLogs:      func() WaspObject { return NewScLogs(o.vm) },
		wasmhost.KeyMaps:      func() WaspObject { return NewScMaps(o.vm) },
		wasmhost.KeyParams:    func() WaspObject { return NewScDictFromKvStore(&o.vm.KvStoreHost, o.vm.params()) },
		wasmhost.KeyResults:   func() WaspObject { return NewScDict(o.vm) },
		wasmhost.KeyState:     func() WaspObject { return NewScDictFromKvStore(&o.vm.KvStoreHost, o.vm.state()) },
		wasmhost.KeyTransfers: func() WaspObject { return NewScTransfers(o.vm) },
		wasmhost.KeyUtility:   func() WaspObject { return NewScUtility(o.vm) },
	})
}

func (o *ScContext) GetTypeId(keyId int32) int32 {
	return typeIds[keyId]
}

func (o *ScContext) SetBytes(keyId int32, typeId int32, bytes []byte) {
	switch keyId {
	case wasmhost.KeyEvent:
		o.vm.ctx.Event(string(bytes))
	case wasmhost.KeyLog:
		o.vm.log().Infof(string(bytes))
	case wasmhost.KeyTrace:
		o.vm.log().Debugf(string(bytes))
	case wasmhost.KeyPanic:
		o.vm.log().Panicf(string(bytes))
	default:
		o.invalidKey(keyId)
	}
}
