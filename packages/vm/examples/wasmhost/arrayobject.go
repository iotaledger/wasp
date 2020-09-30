package wasmhost

import (
	"fmt"
)

type ArrayObject struct {
	vm   *wasmVMPocProcessor
	name string
}

func (a *ArrayObject) checkedObjectId(items *[]int32, index int32, newObject ObjFactory, typeId int32, expectedTypeId int32) int32 {
	if typeId != OBJTYPE_MAP {
		a.error("GetObjectId: Invalid type")
		return 0
	}
	length := int32(len(*items))
	if index < 0 || index > length {
		a.error("GetObjectId: Invalid index")
		return 0
	}
	if index < length {
		return (*items)[index]
	}
	objId := a.vm.AddObject(newObject(a.vm))
	*items = append(*items, objId)
	return objId
}

func (a *ArrayObject) error(format string, args ...interface{}) {
	a.vm.SetError(a.name + "." + fmt.Sprintf(format, args...))
}

func (a *ArrayObject) GetBytes(keyId int32) []byte {
	a.error("GetBytes: Invalid key")
	return []byte(nil)
}

func (a *ArrayObject) GetInt(keyId int32) int64 {
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

func (a *ArrayObject) SetBytes(keyId int32, value []byte) {
	a.error("SetBytes: Immutable")
}

func (a *ArrayObject) SetInt(keyId int32, value int64) {
	a.error("SetInt: Immutable")
}

func (a *ArrayObject) SetString(keyId int32, value string) {
	a.error("SetString: Immutable")
}
