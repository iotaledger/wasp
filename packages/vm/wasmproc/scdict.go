// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"strings"
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
	host      *wasmhost.KvStoreHost
	id        int32
	isMutable bool
	isRoot    bool
	keyId     int32
	kvStore   kv.KVStore
	length    int32
	name      string
	objects   map[int32]int32
	ownerId   int32
	typeId    int32
	types     map[int32]int32
}

func NewScDict(vm *wasmProcessor) *ScDict {
	return NewScDictFromKvStore(&vm.KvStoreHost, dict.New())
}

func NewScDictFromKvStore(host *wasmhost.KvStoreHost, kvStore kv.KVStore) *ScDict {
	o := &ScDict{}
	o.host = host
	o.kvStore = kvStore
	return o
}

func NewNullObject(host *wasmhost.KvStoreHost) WaspObject {
	o := &ScSandboxObject{}
	o.host = host
	o.name = "null"
	o.isRoot = true
	return o
}

func (o *ScDict) InitObj(id int32, keyId int32, owner *ScDict) {
	o.id = id
	o.keyId = keyId
	o.ownerId = owner.id
	o.host = owner.host
	o.isRoot = o.kvStore != nil
	if !o.isRoot {
		o.kvStore = owner.kvStore
	}
	ownerObj := o.Owner()
	o.typeId = ownerObj.GetTypeId(keyId)
	o.name = owner.name + ownerObj.Suffix(keyId)
	if o.ownerId == 1 {
		if strings.HasPrefix(o.name, "root.") {
			// strip off "root." prefix
			o.name = o.name[len("root."):]
		}
		if strings.HasPrefix(o.name, ".") {
			// strip off "." prefix
			o.name = o.name[1:]
		}
	}
	if (o.typeId&wasmhost.OBJTYPE_ARRAY) != 0 && o.kvStore != nil {
		key := o.NestedKey()[1:]
		length, _, err := codec.DecodeInt64(o.kvStore.MustGet(kv.Key(key)))
		if err != nil {
			o.Panic("InitObj: %v", err)
		}
		o.length = int32(length)
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
	objId = o.host.TrackObject(newObject)
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
		return int64(o.length)
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
		o.Panic("MustGetString: %v", err)
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
	return o.host.FindObject(o.ownerId).(WaspObject)
}

func (o *ScDict) Panic(format string, args ...interface{}) {
	err := o.name + "." + fmt.Sprintf(format, args...)
	o.Trace(err)
	panic(err)
}

func (o *ScDict) SetBytes(keyId int32, value []byte) {
	o.validateMutable(keyId)
	//TODO what about AGENT/ADDRESS/COLOR?
	o.kvStore.Set(o.key(keyId, wasmhost.OBJTYPE_BYTES), value)
}

func (o *ScDict) SetInt(keyId int32, value int64) {
	o.validateMutable(keyId)
	switch keyId {
	case wasmhost.KeyLength:
		if o.kvStore != nil {
			//TODO this goes wrong for state, should clear map tree instead
			o.kvStore = dict.New()
			//if (o.typeId & wasmhost.OBJTYPE_ARRAY) != 0 {
			//	key := o.NestedKey()[1:]
			//	o.kvStore.Del(kv.Key(key))
			//}
		}
		o.objects = make(map[int32]int32)
		o.length = 0
	default:
		o.kvStore.Set(o.key(keyId, wasmhost.OBJTYPE_INT), codec.EncodeInt64(value))
	}
}

func (o *ScDict) SetString(keyId int32, value string) {
	o.validateMutable(keyId)
	o.kvStore.Set(o.key(keyId, wasmhost.OBJTYPE_STRING), codec.EncodeString(value))
}

func (o *ScDict) Suffix(keyId int32) string {
	if (o.typeId & wasmhost.OBJTYPE_ARRAY) != 0 {
		return fmt.Sprintf(".%d", keyId)
	}
	key := o.host.GetKeyFromId(keyId)
	if (keyId & wasmhost.KeyFromString) != 0 {
		return "." + string(key)
	}
	return "." + string(key)
}

func (o *ScDict) Trace(format string, a ...interface{}) {
	o.host.Trace(format, a...)
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
		arrayTypeId := o.typeId &^ wasmhost.OBJTYPE_ARRAY
		if typeId == wasmhost.OBJTYPE_BYTES {
			switch arrayTypeId {
			case wasmhost.OBJTYPE_ADDRESS:
			case wasmhost.OBJTYPE_AGENT:
			case wasmhost.OBJTYPE_BYTES:
			case wasmhost.OBJTYPE_COLOR:
			case wasmhost.OBJTYPE_HASH:
			default:
				o.Panic("validate: Invalid byte type")
			}
		} else if arrayTypeId != typeId {
			o.Panic("validate: Invalid type")
		}
		if /*o.isMutable && */ keyId == o.length {
			o.length++
			if o.kvStore != nil {
				key := o.NestedKey()[1:]
				o.kvStore.Set(kv.Key(key), codec.EncodeInt64(int64(o.length)))
			}
			return
		}
		if keyId < 0 || keyId >= o.length {
			o.Panic("validate: Invalid index")
		}
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

func (o *ScDict) validateMutable(keyId int32) {
	//TODO
	//if !o.isMutable {
	//	o.panic("validate: Immutable field: %s key %d", o.name, keyId)
	//}
}
