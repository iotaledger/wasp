// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"fmt"
)

type ObjFactory func() WaspObject
type ObjFactories map[int32]ObjFactory

type WaspObject interface {
	HostObject
	InitObj(id int32, keyId int32, owner *ModelObject)
	Error(format string, args ...interface{})
	FindOrMakeObjectId(keyId int32, factory ObjFactory) int32
	Name() string
	Suffix(keyId int32) string
}

func GetArrayObjectId(arrayObj WaspObject, index int32, typeId int32, factory ObjFactory) int32 {
	if !arrayObj.Exists(index) {
		arrayObj.Error("GetArrayObjectId: Invalid index")
		return 0
	}
	if typeId != arrayObj.GetTypeId(index) {
		arrayObj.Error("GetArrayObjectId: Invalid type")
		return 0
	}
	return arrayObj.FindOrMakeObjectId(index, factory)
}

func GetMapObjectId(mapObj WaspObject, keyId int32, typeId int32, factories ObjFactories) int32 {
	factory, ok := factories[keyId]
	if !ok {
		mapObj.Error("GetMapObjectId: Invalid key")
		return 0
	}
	if typeId != mapObj.GetTypeId(keyId) {
		mapObj.Error("GetMapObjectId: Invalid type")
		return 0
	}
	return mapObj.FindOrMakeObjectId(keyId, factory)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ModelObject struct {
	vm      *wasmProcessor
	id      int32
	keyId   int32
	ownerId int32
}

func NewNullObject(vm *wasmProcessor) WaspObject {
	return &ModelObject{vm: vm, id: 0}
}

func (o *ModelObject) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.id = id
	o.keyId = keyId
	o.ownerId = owner.id
	o.vm = owner.vm
	o.vm.Trace("InitObj %s", o.Name())
}

func (o *ModelObject) Error(format string, args ...interface{}) {
	o.vm.SetError(o.Name() + "." + fmt.Sprintf(format, args...))
}

func (o *ModelObject) Exists(keyId int32) bool {
	o.vm.LogText("IMPLEMENT " + o.Name() + ".Exists???")
	return false
}

func (o *ModelObject) FindOrMakeObjectId(keyId int32, factory ObjFactory) int32 {
	panic("implement me")
}

func (o *ModelObject) GetBytes(keyId int32) []byte {
	o.Error("GetBytes: Invalid key")
	return []byte(nil)
}

func (o *ModelObject) GetInt(keyId int32) int64 {
	o.Error("GetInt: Invalid key")
	return 0
}

func (o *ModelObject) GetObjectId(keyId int32, typeId int32) int32 {
	o.Error("GetObjectId: Invalid key")
	return 0
}

func (o *ModelObject) GetString(keyId int32) string {
	o.Error("GetString: Invalid key")
	return ""
}

func (o *ModelObject) GetTypeId(keyId int32) int32 {
	o.Error("GetTypeId: Invalid key")
	return -1
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

func (o *ModelObject) SetBytes(keyId int32, value []byte) {
	o.Error("SetBytes: Immutable")
}

func (o *ModelObject) SetInt(keyId int32, value int64) {
	o.Error("SetInt: Immutable")
}

func (o *ModelObject) SetString(keyId int32, value string) {
	o.Error("SetString: Immutable")
}

func (o *ModelObject) Suffix(keyId int32) string {
	return "." + string(o.vm.getKeyFromId(keyId))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type MapObject struct {
	ModelObject
	objects map[int32]int32
}

func (o *MapObject) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.ModelObject.InitObj(id, keyId, owner)
	o.objects = make(map[int32]int32)
}

func (o *MapObject) FindOrMakeObjectId(keyId int32, factory ObjFactory) int32 {
	objId, ok := o.objects[keyId]
	if ok {
		return objId
	}
	newObject := factory()
	objId = o.vm.TrackObject(newObject)
	newObject.InitObj(objId, keyId, &o.ModelObject)
	o.objects[keyId] = objId
	return objId
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
	newObject := factory()
	objId := a.vm.TrackObject(newObject)
	newObject.InitObj(objId, keyId, &a.ModelObject)
	a.objects = append(a.objects, objId)
	return objId
}

func (a *ArrayObject) GetBytes(keyId int32) []byte {
	a.Error("GetBytes: Invalid access")
	return []byte(nil)
}

func (a *ArrayObject) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(len(a.objects))
	}
	a.Error("GetInt: Invalid access")
	return 0
}

func (a *ArrayObject) GetObjectId(keyId int32, typeId int32) int32 {
	a.Error("GetObjectId: Invalid access")
	return 0
}

func (a *ArrayObject) GetString(keyId int32) string {
	a.Error("GetString: Invalid access")
	return ""
}

func (a *ArrayObject) Suffix(keyId int32) string {
	return fmt.Sprintf("[%d]", keyId)
}
