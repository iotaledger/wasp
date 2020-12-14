// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ScDict struct {
	ModelObject
	objects map[int32]int32
	types map[int32]int32
}

func NewScDict(dict kv.KVStore, typeId int32) *ScDict {
	o := &ScDict{}
	o.Dict = dict
	o.typeId = typeId
	return o
}

func (o *ScDict) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.ModelObject.InitObj(id, keyId, owner)
	if o.Dict == nil {
		o.Dict = dict.New()
	}
	if o.typeId == 0 {
		o.typeId = OBJTYPE_MAP
	}
	o.objects = make(map[int32]int32)
	o.types = make(map[int32]int32)
}

func (o *ScDict) Exists(keyId int32) bool {
	suffix := o.Suffix(keyId)
	key := o.NestedKey() + suffix
	o.vm.Trace("Exists:%s, key %s", o.Name()+suffix, key)
	return o.Dict.MustHas(kv.Key(key[1:]))
}

func (o *ScDict) FindOrMakeObjectId(keyId int32, factory ObjFactory) int32 {
	objId, ok := o.objects[keyId]
	if ok {
		return objId
	}
	objId = o.MakeObjectId(keyId, factory)
	o.objects[keyId] = objId
	return objId
}

func (o *ScDict) GetBytes(keyId int32) []byte {
	//TODO what about AGENT/ADDRESS/COLOR?
	return o.GetTypedBytes(keyId, OBJTYPE_BYTES)
}

func (o *ScDict) GetInt(keyId int32) int64 {
	bytes := o.GetTypedBytes(keyId, OBJTYPE_INT)
	value, _, err := codec.DecodeInt64(bytes)
	if err != nil {
		o.Panic("GetInt: %v", err)
	}
	return value
}

func (o *ScDict) GetObjectId(keyId int32, typeId int32) int32 {
	o.validate(keyId, typeId)
	if typeId != OBJTYPE_MAP && (typeId&OBJTYPE_ARRAY) == 0 {
		o.Panic("GetObjectId: Invalid type")
	}
	return GetScDictId(o, keyId, typeId, ObjFactories{
		keyId: func() WaspObject { return NewScDict(o.Dict, typeId) },
	})
}

func (o *ScDict) GetString(keyId int32) string {
	bytes := o.GetTypedBytes(keyId, OBJTYPE_STRING)
	value, _, err := codec.DecodeString(bytes)
	if err != nil {
		o.Panic("GetString: %v", err)
	}
	return value
}

func (o *ScDict) GetTypedBytes(keyId int32, typeId int32) []byte {
	o.validate(keyId, typeId)
	suffix := o.Suffix(keyId)
	key := o.NestedKey() + suffix
	o.vm.Trace("GetTypedBytes: %s, key %s", o.Name()+suffix, key)
	return o.Dict.MustGet(kv.Key(key[1:]))
}

func (o *ScDict) GetTypeId(keyId int32) int32 {
	if (o.typeId & OBJTYPE_ARRAY) != 0 {
		return o.typeId &^ OBJTYPE_ARRAY
	}
	//TODO incomplete, currently only contains used field types
	typeId, ok := o.types[keyId]
	if ok {
		return typeId
	}
	return -1
}

func (o *ScDict) SetBytes(keyId int32, value []byte) {
	//TODO what about AGENT/ADDRESS/COLOR?
	o.SetTypedBytes(keyId, OBJTYPE_BYTES, value)
}

func (o *ScDict) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Dict = dict.New()
	default:
		o.SetTypedBytes(keyId, OBJTYPE_INT, codec.EncodeInt64(value))
	}
}

func (o *ScDict) SetString(keyId int32, value string) {
	o.SetTypedBytes(keyId, OBJTYPE_STRING, codec.EncodeString(value))
}

func (o *ScDict) SetTypedBytes(keyId int32, typeId int32, value []byte) {
	o.validate(keyId, typeId)
	suffix := o.Suffix(keyId)
	key := o.NestedKey() + suffix
	o.vm.Trace("SetTypedBytes: %s, key %s", o.Name()+suffix, key)
	o.Dict.Set(kv.Key(key[1:]), value)
}

func (o *ScDict) validate(keyId int32, typeId int32) {
	if (o.typeId&OBJTYPE_ARRAY) != 0 && o.typeId != typeId {
		// actually array
		o.Panic("validate: Invalid type")
	}
	fieldType, ok := o.types[keyId]
	if !ok {
		// first encounter of this key id, register type to make
		// sure that future usages are all using that same type
		o.types[keyId] = typeId
		return
	}
	if fieldType != typeId {
		o.Panic("validate: Invalid access")
	}
}
