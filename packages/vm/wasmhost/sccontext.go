// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

type ScContext struct {
	MapObject
}

func NewScContext(vm *wasmProcessor) *ScContext {
	o := &ScContext{}
	o.id = 1
	o.vm = vm
	o.objects = make(map[int32]int32)
	return o
}

func (o *ScContext) Exists(keyId int32) bool {
	if keyId == KeyExports {
		return o.vm.ctx == nil && o.vm.ctxView == nil
	}
	return o.GetTypeId(keyId) >= 0
}

func (o *ScContext) Finalize() {
	o.objects = make(map[int32]int32)
	o.vm.objIdToObj = o.vm.objIdToObj[:2]
}

func (o *ScContext) GetBytes(keyId int32) []byte {
	switch keyId {
	case KeyCaller:
		id := o.vm.ctx.Caller()
		return id.Bytes()
	}
	return o.MapObject.GetBytes(keyId)
}

func (o *ScContext) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyTimestamp:
		return o.vm.ctx.GetTimestamp()
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScContext) GetObjectId(keyId int32, typeId int32) int32 {
	if keyId == KeyExports && (o.vm.ctx != nil || o.vm.ctxView != nil) {
		// once map has entries (after on_load) this cannot be called any more
		return o.MapObject.GetObjectId(keyId, typeId)
	}

	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		KeyBalances:  func() WaspObject { return &ScBalances{} },
		KeyCalls:     func() WaspObject { return &ScCalls{} },
		KeyContract:  func() WaspObject { return &ScContract{} },
		KeyExports:   func() WaspObject { return &ScExports{} },
		KeyIncoming:  func() WaspObject { return &ScBalances{incoming: true} },
		KeyLogs:      func() WaspObject { return &ScLogs{} },
		KeyParams:    func() WaspObject { return &ScImmutableDict{Dict: o.vm.Params()} },
		KeyPosts:     func() WaspObject { return &ScPosts{} },
		KeyResults:   func() WaspObject { return &ScMutableDict{} },
		KeyState:     func() WaspObject { return &ScState{} },
		KeyTransfers: func() WaspObject { return &ScTransfers{} },
		KeyUtility:   func() WaspObject { return &ScUtility{} },
		KeyViews:     func() WaspObject { return &ScViews{} },
	})
}

func (o *ScContext) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyBalances:
		return OBJTYPE_MAP
	case KeyCalls:
		return OBJTYPE_MAP | OBJTYPE_ARRAY
	case KeyContract:
		return OBJTYPE_MAP
	case KeyExports:
		return OBJTYPE_STRING | OBJTYPE_ARRAY
	case KeyIncoming:
		return OBJTYPE_MAP
	case KeyLogs:
		return OBJTYPE_MAP
	case KeyParams:
		return OBJTYPE_MAP
	case KeyPosts:
		return OBJTYPE_MAP | OBJTYPE_ARRAY
	case KeyResults:
		return OBJTYPE_MAP
	case KeyCaller:
		return OBJTYPE_BYTES
	case KeyState:
		return OBJTYPE_MAP
	case KeyTimestamp:
		return OBJTYPE_INT
	case KeyTransfers:
		return OBJTYPE_MAP | OBJTYPE_ARRAY
	case KeyUtility:
		return OBJTYPE_MAP
	case KeyViews:
		return OBJTYPE_MAP | OBJTYPE_ARRAY
	}
	return -1
}
