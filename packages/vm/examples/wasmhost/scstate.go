package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type ScState struct {
	MapObject
	fields map[int32]int32
	types  map[int32]int32
}

func (o *ScState) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.fields = make(map[int32]int32)
	o.types = make(map[int32]int32)
}

func (o *ScState) GetBytes(keyId int32) []byte {
	if !o.valid(keyId, OBJTYPE_BYTES) {
		return []byte(nil)
	}
	key := kv.Key(o.vm.GetKey(keyId))
	return o.vm.ctx.AccessState().Get(key)
}

func (o *ScState) GetInt(keyId int32) int64 {
	if !o.valid(keyId, OBJTYPE_INT) {
		return 0
	}
	key := kv.Key(o.vm.GetKey(keyId))
	value, _ := o.vm.ctx.AccessState().GetInt64(key)
	return value
}

func (o *ScState) GetObjectId(keyId int32, typeId int32) int32 {
	if !o.valid(keyId, typeId) {
		return 0
	}
	objId, ok := o.fields[keyId]
	if ok {
		return objId
	}
	switch typeId {
	case OBJTYPE_BYTES_ARRAY:
		objId = o.vm.TrackObject(NewScStateArray(o.vm, keyId, OBJTYPE_BYTES))
	case OBJTYPE_INT_ARRAY:
		objId = o.vm.TrackObject(NewScStateArray(o.vm, keyId, OBJTYPE_INT))
	case OBJTYPE_MAP:
		objId = o.vm.TrackObject(NewScStateMap(o.vm, keyId))
	case OBJTYPE_STRING_ARRAY:
		objId = o.vm.TrackObject(NewScStateArray(o.vm, keyId, OBJTYPE_STRING))
	default:
		o.error("GetObjectId: Invalid type id")
		return 0
	}
	o.fields[keyId] = objId
	return objId
}

func (o *ScState) GetString(keyId int32) string {
	if !o.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	key := kv.Key(o.vm.GetKey(keyId))
	value, _ := o.vm.ctx.AccessState().GetString(key)
	return value
}

func (o *ScState) SetBytes(keyId int32, value []byte) {
	if !o.valid(keyId, OBJTYPE_BYTES) {
		return
	}
	key := kv.Key(o.vm.GetKey(keyId))
	o.vm.ctx.AccessState().Set(key, value)
}

func (o *ScState) SetInt(keyId int32, value int64) {
	if !o.valid(keyId, OBJTYPE_INT) {
		return
	}
	key := kv.Key(o.vm.GetKey(keyId))
	o.vm.ctx.AccessState().SetInt64(key, value)
}

func (o *ScState) SetString(keyId int32, value string) {
	if !o.valid(keyId, OBJTYPE_STRING) {
		return
	}
	key := kv.Key(o.vm.GetKey(keyId))
	o.vm.ctx.AccessState().SetString(key, value)
}

func (o *ScState) valid(keyId int32, typeId int32) bool {
	fieldType, ok := o.types[keyId]
	if !ok {
		o.types[keyId] = typeId
		return true
	}
	if fieldType != typeId {
		o.error("valid: Invalid access")
		return false
	}
	return true
}
