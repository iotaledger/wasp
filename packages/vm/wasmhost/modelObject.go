package wasmhost

import (
	"fmt"
)

type MapFactory func() WaspObject
type MapFactories map[int32]MapFactory

type WaspObject interface {
	HostObject
	InitVM(vm *wasmProcessor, keyId int32)
	Error(format string, args ...interface{})
	FindOrMakeObjectId(keyId int32, factory MapFactory) int32
}

func GetArrayObjectId(arrayObj WaspObject, index int32, typeId int32, factory MapFactory) int32 {
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

func GetMapObjectId(mapObj WaspObject, keyId int32, typeId int32, factories MapFactories) int32 {
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
	vm    *wasmProcessor
	keyId int32
	name  string
}

func NewNullObject(vm *wasmProcessor) WaspObject {
	return &ModelObject{vm: vm, name: "null"}
}

func (o *ModelObject) InitVM(vm *wasmProcessor, keyId int32) {
	o.vm = vm
	o.keyId = keyId
}

func (o *ModelObject) Error(format string, args ...interface{}) {
	if o.keyId != 0 {
		o.name = string(o.vm.GetKey(o.keyId))
		o.keyId = 0
	}
	o.vm.SetError(o.name + "." + fmt.Sprintf(format, args...))
}

func (o *ModelObject) Exists(keyId int32) bool {
	o.vm.LogText("IMPLEMENT " + o.name + ".Exists???")
	return false
}

func (o *ModelObject) FindOrMakeObjectId(keyId int32, factory MapFactory) int32 {
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

func (o *ModelObject) SetBytes(keyId int32, value []byte) {
	o.Error("SetBytes: Immutable")
}

func (o *ModelObject) SetInt(keyId int32, value int64) {
	o.Error("SetInt: Immutable")
}

func (o *ModelObject) SetString(keyId int32, value string) {
	o.Error("SetString: Immutable")
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type MapObject struct {
	ModelObject
	objects map[int32]int32
}

func (o *MapObject) InitVM(vm *wasmProcessor, keyId int32) {
	o.ModelObject.InitVM(vm, keyId)
	o.objects = make(map[int32]int32)
}

func (o *MapObject) FindOrMakeObjectId(keyId int32, factory MapFactory) int32 {
	objId, ok := o.objects[keyId]
	if ok {
		return objId
	}
	newObject := factory()
	newObject.InitVM(o.vm, keyId)
	objId = o.vm.TrackObject(newObject)
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

func (a *ArrayObject) FindOrMakeObjectId(keyId int32, factory MapFactory) int32 {
	if keyId < int32(len(a.objects)) {
		return a.objects[keyId]
	}
	newObject := factory()
	newObject.InitVM(a.vm, 0)
	objId := a.vm.TrackObject(newObject)
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
