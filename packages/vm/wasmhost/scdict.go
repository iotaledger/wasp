// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/mr-tron/base58"
)

type ObjFactory func() WaspObject
type ObjFactories map[int32]ObjFactory

type WaspObject interface {
	HostObject
	InitObj(id int32, keyId int32, owner *ScDict)
	Panic(format string, args ...interface{})
	FindOrMakeObjectId(keyId int32, factory ObjFactory) int32
	Name() string
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
	objects map[int32]int32
	ownerId int32
	typeId  int32
	types   map[int32]int32
	vm      *wasmProcessor
}

func NewNullObject(vm *wasmProcessor) WaspObject {
	return &ScDict{vm: vm, id: 0, isRoot: true}
}

func (o *ScDict) InitObj(id int32, keyId int32, owner *ScDict) {
	o.id = id
	o.keyId = keyId
	o.ownerId = owner.id
	o.isRoot = o.kvStore != nil
	if !o.isRoot {
		o.kvStore = owner.kvStore
	}
	o.vm = owner.vm
	o.vm.Trace("InitObj %s", o.Name())
	o.typeId = o.vm.FindObject(owner.id).GetTypeId(keyId)
	o.objects = make(map[int32]int32)
	o.types = make(map[int32]int32)
}

func (o *ScDict) Exists(keyId int32) bool {
	if o.typeId == (OBJTYPE_ARRAY | OBJTYPE_MAP) {
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
	return o.kvStore.MustGet(o.key(keyId, OBJTYPE_BYTES))
}

func (o *ScDict) GetInt(keyId int32) int64 {
	if (o.typeId&OBJTYPE_ARRAY) != 0 && keyId == KeyLength {
		return int64(len(o.objects))
	}
	bytes := o.kvStore.MustGet(o.key(keyId, OBJTYPE_INT))
	value, _, err := codec.DecodeInt64(bytes)
	if err != nil {
		o.Panic("GetInt: %v", err)
	}
	return value
}

func (o *ScDict) GetObjectId(keyId int32, typeId int32) int32 {
	o.validate(keyId, typeId)
	if (typeId&OBJTYPE_ARRAY) == 0 && typeId != OBJTYPE_MAP {
		o.Panic("GetObjectId: Invalid type")
	}
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		keyId: func() WaspObject { return &ScDict{} },
	})
}

func (o *ScDict) GetString(keyId int32) string {
	bytes := o.kvStore.MustGet(o.key(keyId, OBJTYPE_STRING))
	value, _, err := codec.DecodeString(bytes)
	if err != nil {
		o.Panic("GetString: %v", err)
	}
	return value
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
	return 0
}

func (o *ScDict) key(keyId int32, typeId int32) kv.Key {
	o.validate(keyId, typeId)
	suffix := o.Suffix(keyId)
	key := o.NestedKey() + suffix
	o.vm.Trace("fld: %s", o.Name()+suffix)
	o.vm.Trace("key: %s", key[1:])
	return kv.Key(key[1:])
}

func (o *ScDict) Name() string {
	switch o.id {
	case 0:
		return "null"
	case 1:
		return "root"
	default:
		owner := o.vm.objIdToObj[o.ownerId].(WaspObject)
		if o.ownerId == 1 {
			// root sub object, skip the "root." prefix
			return string(o.vm.getKeyFromId(o.keyId))
		}
		return owner.Name() + owner.Suffix(o.keyId)
	}
}

func (o *ScDict) NestedKey() string {
	if o.isRoot {
		return ""
	}
	owner := o.vm.objIdToObj[o.ownerId].(WaspObject)
	return owner.NestedKey() + owner.Suffix(o.keyId)
}

func (o *ScDict) Panic(format string, args ...interface{}) {
	err := o.Name() + "." + fmt.Sprintf(format, args...)
	o.vm.LogText(err)
	panic(err)
}

func (o *ScDict) SetBytes(keyId int32, value []byte) {
	//TODO what about AGENT/ADDRESS/COLOR?
	o.kvStore.Set(o.key(keyId, OBJTYPE_BYTES), value)
}

func (o *ScDict) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		//TODO this goes wrong for state, should clear map instead
		o.kvStore = dict.New()
		o.objects = make(map[int32]int32)
	default:
		o.kvStore.Set(o.key(keyId, OBJTYPE_INT), codec.EncodeInt64(value))
	}
}

func (o *ScDict) SetString(keyId int32, value string) {
	o.kvStore.Set(o.key(keyId, OBJTYPE_STRING), codec.EncodeString(value))
}

func (o *ScDict) Suffix(keyId int32) string {
	if (o.typeId & OBJTYPE_ARRAY) != 0 {
		return fmt.Sprintf(".%d", keyId)
	}
	bytes := o.vm.getKeyFromId(keyId)
	if (keyId & KeyFromString) != 0 {
		return "." + string(bytes)
	}
	return "." + base58.Encode(bytes)
}

func (o *ScDict) validate(keyId int32, typeId int32) {
	if o.kvStore == nil {
		o.Panic("validate: Missing kvstore")
	}
	if typeId == -1 {
		return
	}
	if (o.typeId & OBJTYPE_ARRAY) != 0 {
		// actually array
		if (o.typeId &^ OBJTYPE_ARRAY) != typeId {
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
