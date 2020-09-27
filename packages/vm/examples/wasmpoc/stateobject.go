package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasplib/host/interfaces"
	"github.com/iotaledger/wasplib/host/interfaces/objtype"
)

type StateObject struct {
	MapObject
	fields map[int32]int32
	types  map[int32]int32
}

func NewStateObject(h *wasmVMPocProcessor) interfaces.HostObject {
	return &StateObject{MapObject: MapObject{vm: h, name: "State"}, fields: make(map[int32]int32), types: make(map[int32]int32)}
}

func (o *StateObject) GetBytes(keyId int32) []byte {
	if !o.valid(keyId, objtype.OBJTYPE_BYTES) {
		return []byte(nil)
	}
	key := kv.Key(o.vm.GetKey(keyId))
	return o.vm.ctx.AccessState().Get(key)
}

func (o *StateObject) GetInt(keyId int32) int64 {
	if !o.valid(keyId, objtype.OBJTYPE_INT) {
		return 0
	}
	key := kv.Key(o.vm.GetKey(keyId))
	value, _ := o.vm.ctx.AccessState().GetInt64(key)
	return value
}

func (o *StateObject) GetObjectId(keyId int32, typeId int32) int32 {
	if !o.valid(keyId, typeId) {
		return 0
	}
	objId, ok := o.fields[keyId]
	if ok {
		return objId
	}
	key := kv.Key(o.vm.GetKey(keyId))
	switch typeId {
	case objtype.OBJTYPE_BYTES_ARRAY:
		a := o.vm.ctx.AccessState().GetArray(key)
		objId = o.vm.AddObject(NewStateArray(o.vm, a, objtype.OBJTYPE_BYTES))
	case objtype.OBJTYPE_INT_ARRAY:
		a := o.vm.ctx.AccessState().GetArray(key)
		objId = o.vm.AddObject(NewStateArray(o.vm, a, objtype.OBJTYPE_INT))
	case objtype.OBJTYPE_MAP:
		m := o.vm.ctx.AccessState().GetDictionary(key)
		objId = o.vm.AddObject(NewStateMap(o.vm, m))
	case objtype.OBJTYPE_STRING_ARRAY:
		a := o.vm.ctx.AccessState().GetArray(key)
		objId = o.vm.AddObject(NewStateArray(o.vm, a, objtype.OBJTYPE_STRING))
	default:
		o.vm.SetError("Invalid type id")
		return 0
	}
	o.fields[keyId] = objId
	return objId
}

func (o *StateObject) GetString(keyId int32) string {
	if !o.valid(keyId, objtype.OBJTYPE_STRING) {
		return ""
	}
	key := kv.Key(o.vm.GetKey(keyId))
	value, _ := o.vm.ctx.AccessState().GetString(key)
	return value
}

func (o *StateObject) SetBytes(keyId int32, value []byte) {
	if !o.valid(keyId, objtype.OBJTYPE_BYTES) {
		return
	}
	key := kv.Key(o.vm.GetKey(keyId))
	o.vm.ctx.AccessState().Set(key, value)
}

func (o *StateObject) SetInt(keyId int32, value int64) {
	if !o.valid(keyId, objtype.OBJTYPE_INT) {
		return
	}
	key := kv.Key(o.vm.GetKey(keyId))
	o.vm.ctx.AccessState().SetInt64(key, value)
}

func (o *StateObject) SetString(keyId int32, value string) {
	if !o.valid(keyId, objtype.OBJTYPE_STRING) {
		return
	}
	key := kv.Key(o.vm.GetKey(keyId))
	o.vm.ctx.AccessState().SetString(key, value)
}

func (o *StateObject) valid(keyId int32, typeId int32) bool {
	fieldType, ok := o.types[keyId]
	if !ok {
		o.types[keyId] = typeId
		return true
	}
	if fieldType != typeId {
		o.vm.SetError("Invalid access")
		return false
	}
	return true
}
