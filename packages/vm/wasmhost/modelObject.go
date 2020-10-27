package wasmhost

import (
	"fmt"
)

type WaspObject interface {
	HostObject
	InitVM(vm *wasmProcessor, keyId int32)
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

func (o *ModelObject) error(format string, args ...interface{}) {
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

func (o *ModelObject) GetBytes(keyId int32) []byte {
	o.error("GetBytes: Invalid key")
	return []byte(nil)
}

func (o *ModelObject) GetInt(keyId int32) int64 {
	o.error("GetInt: Invalid key")
	return 0
}

func (o *ModelObject) GetObjectId(keyId int32, typeId int32) int32 {
	o.error("GetObjectId: Invalid key")
	return 0
}

func (o *ModelObject) GetString(keyId int32) string {
	o.error("GetString: Invalid key")
	return ""
}

func (o *ModelObject) SetBytes(keyId int32, value []byte) {
	o.error("SetBytes: Immutable")
}

func (o *ModelObject) SetInt(keyId int32, value int64) {
	o.error("SetInt: Immutable")
}

func (o *ModelObject) SetString(keyId int32, value string) {
	o.error("SetString: Immutable")
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

type MapObjDesc struct {
	typeId  int32
	factory func() WaspObject
}

func (o *MapObject) GetMapObjectId(keyId int32, typeId int32, desc map[int32]MapObjDesc) int32 {
	objDesc, ok := desc[keyId]
	if !ok {
		return o.GetObjectId(keyId, typeId)
	}
	if typeId != objDesc.typeId {
		o.error("GetObjectId: Invalid type")
		return 0
	}
	objId, ok := o.objects[keyId]
	if ok {
		return objId
	}
	newObject := objDesc.factory()
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

type ObjFactory func(vm *wasmProcessor) HostObject

func (a *ArrayObject) GetArrayObjectId(index int32, typeId int32, factory func() WaspObject) int32 {
	if typeId != OBJTYPE_MAP {
		a.error("GetObjectId: Invalid type")
		return 0
	}
	length := int32(len(a.objects))
	if index < 0 || index > length {
		a.error("GetObjectId: Invalid index")
		return 0
	}
	if index < length {
		return a.objects[index]
	}
	newObject := factory()
	newObject.InitVM(a.vm, 0)
	objId := a.vm.TrackObject(newObject)
	a.objects = append(a.objects, objId)
	return objId
}

func (a *ArrayObject) GetBytes(keyId int32) []byte {
	a.error("GetBytes: Invalid access")
	return []byte(nil)
}

func (a *ArrayObject) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(len(a.objects))
	}
	a.error("GetInt: Invalid access")
	return 0
}

func (a *ArrayObject) GetObjectId(keyId int32, typeId int32) int32 {
	a.error("GetObjectId: Invalid access")
	return 0
}

func (a *ArrayObject) GetString(keyId int32) string {
	a.error("GetString: Invalid access")
	return ""
}
