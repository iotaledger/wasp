// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

const (
	KeyAgent       = KeyUserDefined
	KeyAmount      = KeyAgent - 1
	KeyBalances    = KeyAmount - 1
	KeyBase58      = KeyBalances - 1
	KeyCaller      = KeyBase58 - 1
	KeyCalls       = KeyCaller - 1
	KeyChain       = KeyCalls - 1
	KeyColor       = KeyChain - 1
	KeyContract    = KeyColor - 1
	KeyData        = KeyContract - 1
	KeyDelay       = KeyData - 1
	KeyDescription = KeyDelay - 1
	KeyExports     = KeyDescription - 1
	KeyFunction    = KeyExports - 1
	KeyHash        = KeyFunction - 1
	KeyId          = KeyHash - 1
	KeyIncoming    = KeyId - 1
	KeyIota        = KeyIncoming - 1
	KeyLogs        = KeyIota - 1
	KeyName        = KeyLogs - 1
	KeyOwner       = KeyName - 1
	KeyParams      = KeyOwner - 1
	KeyPosts       = KeyParams - 1
	KeyRandom      = KeyPosts - 1
	KeyResults     = KeyRandom - 1
	KeyState       = KeyResults - 1
	KeyTimestamp   = KeyState - 1
	KeyTransfers   = KeyTimestamp - 1
	KeyUtility     = KeyTransfers - 1
	KeyViews       = KeyUtility - 1
)

var keyMap = map[string]int32{
	// predefined keys
	"error":     KeyError,
	"length":    KeyLength,
	"log":       KeyLog,
	"trace":     KeyTrace,
	"traceHost": KeyTraceHost,
	"warning":   KeyWarning,

	// user-defined keys
	"agent":       KeyAgent,
	"amount":      KeyAmount,
	"balances":    KeyBalances,
	"base58":      KeyBase58,
	"caller":      KeyCaller,
	"calls":       KeyCalls,
	"chain":       KeyChain,
	"color":       KeyColor,
	"contract":    KeyContract,
	"data":        KeyData,
	"delay":       KeyDelay,
	"description": KeyDescription,
	"exports":     KeyExports,
	"function":    KeyFunction,
	"hash":        KeyHash,
	"id":          KeyId,
	"incoming":    KeyIncoming,
	"iota":        KeyIota,
	"logs":        KeyLogs,
	"name":        KeyName,
	"owner":       KeyOwner,
	"params":      KeyParams,
	"posts":       KeyPosts,
	"random":      KeyRandom,
	"results":     KeyResults,
	"state":       KeyState,
	"timestamp":   KeyTimestamp,
	"transfers":   KeyTransfers,
	"utility":     KeyUtility,
	"views":       KeyViews,
}

type ScContext struct {
	MapObject
}

func NewScContext(vm *wasmProcessor) *ScContext {
	return &ScContext{MapObject: MapObject{ModelObject: ModelObject{vm: vm, id: 1}, objects: make(map[int32]int32)}}
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
		// once map has entries (onLoad) this cannot be called any more
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
		return OBJTYPE_MAP_ARRAY
	case KeyContract:
		return OBJTYPE_MAP
	case KeyExports:
		return OBJTYPE_STRING_ARRAY
	case KeyIncoming:
		return OBJTYPE_MAP
	case KeyLogs:
		return OBJTYPE_MAP
	case KeyParams:
		return OBJTYPE_MAP
	case KeyPosts:
		return OBJTYPE_MAP_ARRAY
	case KeyResults:
		return OBJTYPE_MAP
	case KeyCaller:
		return OBJTYPE_BYTES
	case KeyState:
		return OBJTYPE_MAP
	case KeyTimestamp:
		return OBJTYPE_INT
	case KeyTransfers:
		return OBJTYPE_MAP_ARRAY
	case KeyUtility:
		return OBJTYPE_MAP
	case KeyViews:
		return OBJTYPE_MAP_ARRAY
	}
	return -1
}
