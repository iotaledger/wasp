package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

type ScStateMap struct {
	MapObject
	items *kv.MustDictionary
	types map[int32]int32
}

func NewScStateMap(vm *wasmProcessor, keyId int32) HostObject {
	key := vm.GetKey(keyId)
	items := vm.ctx.AccessState().GetDictionary(kv.Key(key))
	return &ScStateMap{MapObject: MapObject{ModelObject: ModelObject{vm: vm, name: "atate.map." + key}}, items: items, types: make(map[int32]int32)}
}

func (m *ScStateMap) GetBytes(keyId int32) []byte {
	if !m.valid(keyId, OBJTYPE_BYTES) {
		return []byte(nil)
	}
	key := []byte(m.vm.GetKey(keyId))
	return m.items.GetAt(key)
}

func (m *ScStateMap) GetInt(keyId int32) int64 {
	if !m.valid(keyId, OBJTYPE_INT) {
		return 0
	}
	key := []byte(m.vm.GetKey(keyId))
	value, _ := kv.DecodeInt64(m.items.GetAt(key))
	return value
}

func (m *ScStateMap) GetObjectId(keyId int32, typeId int32) int32 {
	m.error("GetObjectId: Invalid access")
	return 0
}

func (m *ScStateMap) GetString(keyId int32) string {
	if !m.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	key := []byte(m.vm.GetKey(keyId))
	return string(m.items.GetAt(key))
}

func (m *ScStateMap) SetBytes(keyId int32, value []byte) {
	if !m.valid(keyId, OBJTYPE_BYTES) {
		return
	}
	key := []byte(m.vm.GetKey(keyId))
	m.items.SetAt(key, value)
}

func (m *ScStateMap) SetInt(keyId int32, value int64) {
	if keyId == KeyLength {
		m.error("SetInt: Invalid clear")
		return
	}
	if !m.valid(keyId, OBJTYPE_INT) {
		return
	}
	key := []byte(m.vm.GetKey(keyId))
	m.items.SetAt(key, util.Uint64To8Bytes(uint64(value)))
}

func (m *ScStateMap) SetString(keyId int32, value string) {
	if !m.valid(keyId, OBJTYPE_STRING) {
		return
	}
	key := []byte(m.vm.GetKey(keyId))
	m.items.SetAt(key, []byte(value))
}

func (m *ScStateMap) valid(keyId int32, typeId int32) bool {
	fieldType, ok := m.types[keyId]
	if !ok {
		m.types[keyId] = typeId
		return true
	}
	if fieldType != typeId {
		m.error("valid: Invalid access")
		return false
	}
	return true
}
