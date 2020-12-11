// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ScImmutableDict struct {
	MapObject
	Dict  kv.KVStore
	types map[int32]int32
}

func (o *ScImmutableDict) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.MapObject.InitObj(id, keyId, owner)
	if o.Dict == nil {
		o.Dict = dict.New()
	}
	o.types = make(map[int32]int32)
}

func (o *ScImmutableDict) Exists(keyId int32) bool {
	key := o.vm.GetKey(keyId)
	return o.Dict.MustHas(key)
}

func (o *ScImmutableDict) GetBytes(keyId int32) []byte {
	return o.GetTypedBytes(keyId, OBJTYPE_BYTES)
}

func (o *ScImmutableDict) GetInt(keyId int32) int64 {
	bytes := o.GetTypedBytes(keyId, OBJTYPE_INT)
	value, _, err := codec.DecodeInt64(bytes)
	if err != nil {
		o.Panic("GetInt: %v", err)
	}
	return value
}

func (o *ScImmutableDict) GetString(keyId int32) string {
	bytes := o.GetTypedBytes(keyId, OBJTYPE_STRING)
	value, _, err := codec.DecodeString(bytes)
	if err != nil {
		o.Panic("GetString: %v", err)
	}
	return value
}

func (o *ScImmutableDict) GetTypedBytes(keyId int32, typeId int32) []byte {
	o.validate(keyId, typeId)
	key := o.vm.GetKey(keyId)
	return o.Dict.MustGet(key)
}

//TODO incomplete, only contains used field types
func (o *ScImmutableDict) GetTypeId(keyId int32) int32 {
	typeId, ok := o.types[keyId]
	if ok {
		return typeId
	}
	return -1
}

func (o *ScImmutableDict) validate(keyId int32, typeId int32) {
	fieldType, ok := o.types[keyId]
	if !ok {
		// first encounter of this key id, register type to make
		// sure that future usages are all using that same type
		o.types[keyId] = typeId
		return
	}
	if fieldType != typeId {
		o.Panic("valid: Invalid access")
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableDict struct {
	ScImmutableDict
}

func (o *ScMutableDict) SetBytes(keyId int32, value []byte) {
	o.SetTypedBytes(keyId, OBJTYPE_BYTES, value)
}

func (o *ScMutableDict) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Dict = dict.New()
	default:
		o.SetTypedBytes(keyId, OBJTYPE_INT, codec.EncodeInt64(value))
	}
}

func (o *ScMutableDict) SetString(keyId int32, value string) {
	o.SetTypedBytes(keyId, OBJTYPE_STRING, codec.EncodeString(value))
}

func (o *ScMutableDict) SetTypedBytes(keyId int32, typeId int32, value []byte) {
	o.validate(keyId, typeId)
	key := o.vm.GetKey(keyId)
	o.Dict.Set(key, value)
}
