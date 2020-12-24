// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/mr-tron/base58"
)

type ObjFactory func() WaspObject
type ObjFactories map[int32]ObjFactory

type WaspObject interface {
	wasmhost.HostObject
	InitObj(id int32, keyId int32, owner *ScDict)
	Panic(format string, args ...interface{})
	FindOrMakeObjectId(keyId int32, factory ObjFactory) int32
	NestedKey() string
	Suffix(keyId int32) string
}

func GetArrayObjectId(arrayObj WaspObject, index int32, typeId int32, factory ObjFactory) int32 {
	if !arrayObj.Exists(index) {
		arrayObj.Panic("GetArrayObjectId: Invalid index")
	}
	if typeId != arrayObj.GetTypeId(index) {
		arrayObj.Panic("GetArrayObjectId: Invalid type")
	}
	return arrayObj.FindOrMakeObjectId(index, factory)
}

func GetMapObjectId(mapObj WaspObject, keyId int32, typeId int32, factories ObjFactories) int32 {
	factory, ok := factories[keyId]
	if !ok {
		mapObj.Panic("GetMapObjectId: Invalid key")
	}
	if typeId != mapObj.GetTypeId(keyId) {
		mapObj.Panic("GetMapObjectId: Invalid type")
	}
	return mapObj.FindOrMakeObjectId(keyId, factory)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScDict struct {
	id      int32
	isRoot  bool
	keyId   int32
	kvStore kv.KVStore
	name    string
	objects map[int32]int32
	ownerId int32
	typeId  int32
	types   map[int32]int32
	vm      *wasmProcessor
}

func NewNullObject(vm *wasmProcessor) WaspObject {
	return &ScDict{vm: vm, id: 0, isRoot: true, name: "null"}
}

func (o *ScDict) InitObj(id int32, keyId int32, owner *ScDict) {
	o.id = id
	o.keyId = keyId
	o.ownerId = owner.id
	o.vm = owner.vm
	o.isRoot = o.kvStore != nil
	if !o.isRoot {
		o.kvStore = owner.kvStore
	}
	ownerObj := o.Owner()
	o.typeId = ownerObj.GetTypeId(keyId)
	o.name = owner.name + ownerObj.Suffix(keyId)
	if o.ownerId == 1 {
		// strip off "root." prefix
		o.name = o.name[5:]
	}
	o.Trace("InitObj %s", o.name)
	o.objects = make(map[int32]int32)
	o.types = make(map[int32]int32)
}

func (o *ScDict) Exists(keyId int32) bool {
	if o.typeId == (wasmhost.OBJTYPE_ARRAY | wasmhost.OBJTYPE_MAP) {
		return uint32(keyId) <= uint32(len(o.objects))
	}
	return o.kvStore.MustHas(o.key(keyId, -1))
}

func (o *ScDict) FindOrMakeObjectId(keyId int32, factory ObjFactory) int32 {
	objId, ok := o.objects[keyId]
	if ok {
		return objId
	}
	newObject := factory()
	objId = o.vm.TrackObject(newObject)
	newObject.InitObj(objId, keyId, o)
	o.objects[keyId] = objId
	return objId
}

func (o *ScDict) GetBytes(keyId int32) []byte {
	//TODO what about AGENT/ADDRESS/COLOR?
	return o.kvStore.MustGet(o.key(keyId, wasmhost.OBJTYPE_BYTES))
}

func (o *ScDict) GetInt(keyId int32) int64 {
	if (o.typeId&wasmhost.OBJTYPE_ARRAY) != 0 && keyId == wasmhost.KeyLength {
		return int64(len(o.objects))
	}
	bytes := o.kvStore.MustGet(o.key(keyId, wasmhost.OBJTYPE_INT))
	value, _, err := codec.DecodeInt64(bytes)
	if err != nil {
		o.Panic("GetInt: %v", err)
	}
	return value
}

func (o *ScDict) GetObjectId(keyId int32, typeId int32) int32 {
	o.validate(keyId, typeId)
	if (typeId&wasmhost.OBJTYPE_ARRAY) == 0 && typeId != wasmhost.OBJTYPE_MAP {
		o.Panic("GetObjectId: Invalid type")
	}
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		keyId: func() WaspObject { return &ScDict{} },
	})
}

func (o *ScDict) GetString(keyId int32) string {
	bytes := o.kvStore.MustGet(o.key(keyId, wasmhost.OBJTYPE_STRING))
	value, _, err := codec.DecodeString(bytes)
	if err != nil {
		o.Panic("GetString: %v", err)
	}
	return value
}

func (o *ScDict) GetTypeId(keyId int32) int32 {
	if (o.typeId & wasmhost.OBJTYPE_ARRAY) != 0 {
		return o.typeId &^ wasmhost.OBJTYPE_ARRAY
	}
	//TODO incomplete, currently only contains used field types
	typeId, ok := o.types[keyId]
	if ok {
		return typeId
	}
	return 0
}

func (o *ScDict) key(keyId int32, typeId int32) kv.Key {
	o.validate(keyId, typeId)
	suffix := o.Suffix(keyId)
	key := o.NestedKey() + suffix
	o.Trace("fld: %s%s", o.name, suffix)
	o.Trace("key: %s", key[1:])
	return kv.Key(key[1:])
}

func (o *ScDict) NestedKey() string {
	if o.isRoot {
		return ""
	}
	ownerObj := o.Owner()
	return ownerObj.NestedKey() + ownerObj.Suffix(o.keyId)
}

func (o *ScDict) Owner() WaspObject {
	return o.vm.FindObject(o.ownerId).(WaspObject)
}

func (o *ScDict) Panic(format string, args ...interface{}) {
	err := o.name + "." + fmt.Sprintf(format, args...)
	o.Trace(err)
	panic(err)
}

func (o *ScDict) SetBytes(keyId int32, value []byte) {
	//TODO what about AGENT/ADDRESS/COLOR?
	o.kvStore.Set(o.key(keyId, wasmhost.OBJTYPE_BYTES), value)
}

func (o *ScDict) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		//TODO this goes wrong for state, should clear map instead
		o.kvStore = dict.New()
		o.objects = make(map[int32]int32)
	default:
		o.kvStore.Set(o.key(keyId, wasmhost.OBJTYPE_INT), codec.EncodeInt64(value))
	}
}

func (o *ScDict) SetString(keyId int32, value string) {
	o.kvStore.Set(o.key(keyId, wasmhost.OBJTYPE_STRING), codec.EncodeString(value))
}

func (o *ScDict) Suffix(keyId int32) string {
	if (o.typeId & wasmhost.OBJTYPE_ARRAY) != 0 {
		return fmt.Sprintf(".%d", keyId)
	}
	bytes := o.vm.GetKeyFromId(keyId)
	if (keyId & wasmhost.KeyFromString) != 0 {
		return "." + string(bytes)
	}
	return "." + base58.Encode(bytes)
}

func (o *ScDict) Trace(format string, a ...interface{}) {
	o.vm.Trace(format, a...)
}

func (o *ScDict) validate(keyId int32, typeId int32) {
	if o.kvStore == nil {
		o.Panic("validate: Missing kvstore")
	}
	if typeId == -1 {
		return
	}
	if (o.typeId & wasmhost.OBJTYPE_ARRAY) != 0 {
		// actually array
		if (o.typeId &^ wasmhost.OBJTYPE_ARRAY) != typeId {
			o.Panic("validate: Invalid type")
		}
		//TODO validate keyId >=0 && <= length
		return
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
