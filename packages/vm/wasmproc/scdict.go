// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type (
	ObjFactory   func() WaspObject
	ObjFactories map[int32]ObjFactory
)

type WaspObject interface {
	wasmhost.HostObject
	InitObj(id int32, keyID int32, owner *ScDict)
	Panicf(format string, args ...interface{})
	FindOrMakeObjectID(keyID int32, factory ObjFactory) int32
	NestedKey() string
	Suffix(keyID int32) string
}

func GetArrayObjectID(arrayObj WaspObject, index, typeID int32, factory ObjFactory) int32 {
	if !arrayObj.Exists(index, typeID) {
		arrayObj.Panicf("GetArrayObjectID: invalid index")
	}
	if typeID != arrayObj.GetTypeID(index) {
		arrayObj.Panicf("GetArrayObjectID: invalid type")
	}
	return arrayObj.FindOrMakeObjectID(index, factory)
}

func GetMapObjectID(mapObj WaspObject, keyID, typeID int32, factories ObjFactories) int32 {
	factory, ok := factories[keyID]
	if !ok {
		mapObj.Panicf("GetMapObjectID: invalid key")
	}
	if typeID != mapObj.GetTypeID(keyID) {
		mapObj.Panicf("GetMapObjectID: invalid type")
	}
	return mapObj.FindOrMakeObjectID(keyID, factory)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScDict struct {
	host    *wasmhost.KvStoreHost
	id      int32
	isRoot  bool
	keyID   int32
	kvStore kv.KVStore
	length  int32
	name    string
	objects map[int32]int32
	ownerID int32
	typeID  int32
	types   map[int32]int32
}

// TODO iterate over maps

var _ WaspObject = &ScDict{}

func NewScDict(host *wasmhost.KvStoreHost, kvStore kv.KVStore) *ScDict {
	return &ScDict{host: host, kvStore: kvStore}
}

func NewNullObject(host *wasmhost.KvStoreHost) WaspObject {
	o := &ScSandboxObject{}
	o.host = host
	o.name = "null"
	o.isRoot = true
	return o
}

func (o *ScDict) InitObj(id, keyID int32, owner *ScDict) {
	o.id = id
	o.keyID = keyID
	o.ownerID = owner.id
	o.host = owner.host
	o.isRoot = o.kvStore != nil
	if !o.isRoot {
		o.kvStore = owner.kvStore
	}
	ownerObj := o.Owner()
	o.typeID = ownerObj.GetTypeID(keyID)
	o.name = owner.name + ownerObj.Suffix(keyID)
	if o.ownerID == 1 {
		// strip off "root." prefix
		o.name = strings.TrimPrefix(o.name, "root.")
		// strip off "." prefix
		o.name = strings.TrimPrefix(o.name, ".")
	}
	if (o.typeID&wasmhost.OBJTYPE_ARRAY) != 0 && o.kvStore != nil {
		err := o.getArrayLength()
		if err != nil {
			o.Panicf("InitObj: %v", err)
		}
	}
	o.Tracef("InitObj %s", o.name)
	o.objects = make(map[int32]int32)
	o.types = make(map[int32]int32)
}

func (o *ScDict) CallFunc(keyID int32, params []byte) []byte {
	o.Panicf("CallFunc: invalid call")
	return nil
}

func (o *ScDict) DelKey(keyID, typeID int32) {
	if (o.typeID & wasmhost.OBJTYPE_ARRAY) != 0 {
		o.Panicf("DelKey: cannot delete array")
	}
	if o.typeID == wasmhost.OBJTYPE_MAP {
		o.Panicf("DelKey: cannot delete map")
	}
	o.kvStore.Del(o.key(keyID, typeID))
}

func (o *ScDict) Exists(keyID, typeID int32) bool {
	if keyID == wasmhost.KeyLength && (o.typeID&wasmhost.OBJTYPE_ARRAY) != 0 {
		return true
	}
	if o.typeID == (wasmhost.OBJTYPE_ARRAY | wasmhost.OBJTYPE_MAP) {
		return uint32(keyID) <= uint32(len(o.objects))
	}
	ret, _ := o.kvStore.Has(o.key(keyID, typeID))
	return ret
}

func (o *ScDict) FindOrMakeObjectID(keyID int32, factory ObjFactory) int32 {
	objID, ok := o.objects[keyID]
	if ok {
		return objID
	}
	newObject := factory()
	objID = o.host.TrackObject(newObject)
	newObject.InitObj(objID, keyID, o)
	o.objects[keyID] = objID
	if (o.typeID & wasmhost.OBJTYPE_ARRAY) != 0 {
		o.length++
		if o.kvStore != nil {
			key := o.NestedKey()[1:]
			o.kvStore.Set(kv.Key(key), codec.EncodeInt32(o.length))
		}
	}
	return objID
}

func (o *ScDict) getArrayLength() (err error) {
	key := o.NestedKey()[1:]
	bytes := o.kvStore.MustGet(kv.Key(key))
	if (o.typeID & wasmhost.OBJTYPE_ARRAY16) != wasmhost.OBJTYPE_ARRAY16 {
		o.length, err = codec.DecodeInt32(bytes, 0)
		return err
	}

	var length uint16
	length, err = codec.DecodeUint16(bytes, 0)
	o.length = int32(length)
	return err
}

func (o *ScDict) GetBytes(keyID, typeID int32) []byte {
	if keyID == wasmhost.KeyLength && (o.typeID&wasmhost.OBJTYPE_ARRAY) != 0 {
		return codec.EncodeInt32(o.length)
	}
	bytes := o.kvStore.MustGet(o.key(keyID, typeID))
	o.host.TypeCheck(typeID, bytes)
	return bytes
}

func (o *ScDict) GetObjectID(keyID, typeID int32) int32 {
	o.validate(keyID, typeID)
	if (typeID&wasmhost.OBJTYPE_ARRAY) == 0 && typeID != wasmhost.OBJTYPE_MAP {
		o.Panicf("GetObjectID: invalid type")
	}
	return GetMapObjectID(o, keyID, typeID, ObjFactories{
		keyID: func() WaspObject { return &ScDict{} },
	})
}

func (o *ScDict) GetTypeID(keyID int32) int32 {
	if (o.typeID & wasmhost.OBJTYPE_ARRAY) != 0 {
		return o.typeID & wasmhost.OBJTYPE_TYPEMASK
	}
	// TODO incomplete, currently only contains used field types
	typeID, ok := o.types[keyID]
	if ok {
		return typeID
	}
	return 0
}

func (o *ScDict) InvalidKey(keyID int32) {
	o.Panicf("invalid key: %d", keyID)
}

func (o *ScDict) key(keyID, typeID int32) kv.Key {
	o.validate(keyID, typeID)
	suffix := o.Suffix(keyID)
	key := o.NestedKey() + suffix
	o.Tracef("fld: %s%s", o.name, suffix)
	o.Tracef("key: %s", key[1:])
	return kv.Key(key[1:])
}

func (o *ScDict) KvStore() kv.KVStore {
	return o.kvStore
}

func (o *ScDict) NestedKey() string {
	if o.isRoot {
		return ""
	}
	ownerObj := o.Owner()
	return ownerObj.NestedKey() + ownerObj.Suffix(o.keyID)
}

func (o *ScDict) Owner() WaspObject {
	return o.host.FindObject(o.ownerID).(WaspObject)
}

func (o *ScDict) Panicf(format string, args ...interface{}) {
	o.host.Panicf(o.name+"."+format, args...)
}

func (o *ScDict) SetBytes(keyID, typeID int32, bytes []byte) {
	if keyID == wasmhost.KeyLength {
		if o.kvStore != nil {
			// TODO this goes wrong for state, should clear map tree instead
			// o.kvStore = dict.New()
			if (o.typeID & wasmhost.OBJTYPE_ARRAY) != 0 {
				key := o.NestedKey()[1:]
				o.kvStore.Set(kv.Key(key), codec.EncodeInt32(0))
			}
		}
		o.objects = make(map[int32]int32)
		o.length = 0
		return
	}

	key := o.key(keyID, typeID)
	o.host.TypeCheck(typeID, bytes)
	o.kvStore.Set(key, bytes)
}

func (o *ScDict) Suffix(keyID int32) string {
	if (o.typeID & wasmhost.OBJTYPE_ARRAY16) != 0 {
		if (o.typeID & wasmhost.OBJTYPE_ARRAY16) != wasmhost.OBJTYPE_ARRAY16 {
			return fmt.Sprintf(".%d", keyID)
		}

		buf := make([]byte, 3)
		buf[0] = '#'
		binary.LittleEndian.PutUint16(buf[1:], uint16(keyID))
		return string(buf)
	}

	key := o.host.GetKeyFromID(keyID)
	return "." + string(key)
}

func (o *ScDict) Tracef(format string, a ...interface{}) {
	o.host.Tracef(format, a...)
}

func (o *ScDict) validate(keyID, typeID int32) {
	if o.kvStore == nil {
		o.Panicf("validate: Missing kvstore")
	}
	if typeID == -1 {
		return
	}
	if (o.typeID & wasmhost.OBJTYPE_ARRAY) != 0 {
		// actually array
		arrayTypeID := o.typeID & wasmhost.OBJTYPE_TYPEMASK
		if typeID == wasmhost.OBJTYPE_BYTES {
			switch arrayTypeID {
			case wasmhost.OBJTYPE_ADDRESS:
			case wasmhost.OBJTYPE_AGENT_ID:
			case wasmhost.OBJTYPE_BYTES:
			case wasmhost.OBJTYPE_COLOR:
			case wasmhost.OBJTYPE_HASH:
			default:
				o.Panicf("validate: Invalid byte type")
			}
		} else if arrayTypeID != typeID {
			o.Panicf("validate: Invalid type")
		}
		if keyID == o.length {
			switch o.kvStore.(type) {
			case *ScViewState:
				break
			default:
				o.length++
				key := o.NestedKey()[1:]
				o.kvStore.Set(kv.Key(key), codec.EncodeInt32(o.length))
				return
			}
		}
		if keyID < 0 || keyID >= o.length {
			o.Panicf("validate: Invalid index")
		}
		return
	}
	fieldType, ok := o.types[keyID]
	if !ok {
		// first encounter of this key id, register type to make
		// sure that future usages are all using that same type
		o.types[keyID] = typeID
		return
	}
	if fieldType != typeID {
		o.Panicf("validate: Invalid access")
	}
}

func (o *ScDict) SetKvStore(res kv.KVStore) {
	o.kvStore = res
	o.objects = make(map[int32]int32)
	o.types = make(map[int32]int32)
}
