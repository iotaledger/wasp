// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/mr-tron/base58"
)

type ObjFactory func() WaspObject
type ObjFactories map[int32]ObjFactory

type WaspObject interface {
	HostObject
	InitObj(id int32, keyId int32, owner *ModelObject)
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

func GetScDictId(mapObj WaspObject, keyId int32, typeId int32, factories ObjFactories) int32 {
	factory, ok := factories[keyId]
	if !ok {
		mapObj.Panic("GetScDictId: Invalid key")
	}
	if typeId != mapObj.GetTypeId(keyId) {
		mapObj.Panic("GetScDictId: Invalid type")
	}
	return mapObj.FindOrMakeObjectId(keyId, factory)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ModelObject struct {
	vm      *wasmProcessor
	id      int32
	isRoot  bool
	keyId   int32
	ownerId int32
	typeId  int32
	Dict    kv.KVStore
}

func NewNullObject(vm *wasmProcessor) WaspObject {
	return &ModelObject{vm: vm, id: 0}
}

func (o *ModelObject) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.id = id
	o.keyId = keyId
	o.ownerId = owner.id
	if owner.id == 1 {
		o.isRoot = true
	}
	o.vm = owner.vm
	o.vm.Trace("InitObj %s", o.Name())
}

func (o *ModelObject) Exists(keyId int32) bool {
	o.vm.LogText("IMPLEMENT " + o.Name() + ".Exists???")
	return false
}

func (o *ModelObject) FindOrMakeObjectId(keyId int32, factory ObjFactory) int32 {
	o.Panic("implement me")
	return 0
}

func (o *ModelObject) GetBytes(keyId int32) []byte {
	o.Panic("GetBytes: Invalid key")
	return []byte(nil)
}

func (o *ModelObject) GetInt(keyId int32) int64 {
	o.Panic("GetInt: Invalid key")
	return 0
}

func (o *ModelObject) GetObjectId(keyId int32, typeId int32) int32 {
	o.Panic("GetObjectId: Invalid key")
	return 0
}

func (o *ModelObject) GetString(keyId int32) string {
	o.Panic("GetString: Invalid key")
	return ""
}

func (o *ModelObject) GetTypeId(keyId int32) int32 {
	o.Panic("GetTypeId: Invalid key")
	return -1
}

func (o *ModelObject) MakeObjectId(keyId int32, factory ObjFactory) int32 {
	newObject := factory()
	objId := o.vm.TrackObject(newObject)
	newObject.InitObj(objId, keyId, o)
	return objId
}

func (o *ModelObject) Name() string {
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

func (o *ModelObject) NestedKey() string {
	if o.isRoot {
		return ""
	}
	owner := o.vm.objIdToObj[o.ownerId].(WaspObject)
	return owner.NestedKey() + owner.Suffix(o.keyId)
}

func (o *ModelObject) Panic(format string, args ...interface{}) {
	err := o.Name() + "." + fmt.Sprintf(format, args...)
	o.vm.LogText(err)
	panic(err)
}

func (o *ModelObject) SetBytes(keyId int32, value []byte) {
	o.Panic("SetBytes: Immutable")
}

func (o *ModelObject) SetInt(keyId int32, value int64) {
	o.Panic("SetInt: Immutable")
}

func (o *ModelObject) SetString(keyId int32, value string) {
	o.Panic("SetString: Immutable")
}

func (o *ModelObject) Suffix(keyId int32) string {
	if (o.typeId&OBJTYPE_ARRAY) != 0 {
		return fmt.Sprintf("#%d", keyId)
	}
	bytes := o.vm.getKeyFromId(keyId)
	if (keyId & KeyFromString) != 0 {
		return "." + string(bytes)
	}
	return "." + base58.Encode(bytes)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ArrayObject struct {
	ModelObject
	objects []int32
}

func (a *ArrayObject) Exists(keyId int32) bool {
	return uint32(keyId) <= uint32(len(a.objects))
}

func (a *ArrayObject) FindOrMakeObjectId(keyId int32, factory ObjFactory) int32 {
	if keyId < int32(len(a.objects)) {
		return a.objects[keyId]
	}
	objId := a.MakeObjectId(keyId, factory)
	a.objects = append(a.objects, objId)
	return objId
}

func (a *ArrayObject) GetBytes(keyId int32) []byte {
	a.Panic("GetBytes: Invalid access")
	return []byte(nil)
}

func (a *ArrayObject) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(len(a.objects))
	}
	a.Panic("GetInt: Invalid access")
	return 0
}

func (a *ArrayObject) GetObjectId(keyId int32, typeId int32) int32 {
	a.Panic("GetObjectId: Invalid access")
	return 0
}

func (a *ArrayObject) GetString(keyId int32) string {
	a.Panic("GetString: Invalid access")
	return ""
}

func (a *ArrayObject) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return OBJTYPE_MAP
	}
	return -1
}
