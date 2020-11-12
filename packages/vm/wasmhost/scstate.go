package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/util"
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

func (o *ScState) Exists(keyId int32) bool {
	key := o.vm.GetKey(keyId)
	return o.vm.ctx.AccessState().Has(key)
}

func (o *ScState) GetBytes(keyId int32) []byte {
	if !o.valid(keyId, OBJTYPE_BYTES) {
		return []byte(nil)
	}
	key := o.vm.GetKey(keyId)
	return o.vm.ctx.AccessState().Get(key)
}

func (o *ScState) GetInt(keyId int32) int64 {
	if !o.valid(keyId, OBJTYPE_INT) {
		return 0
	}
	key := o.vm.GetKey(keyId)
	value, _ := o.vm.ctx.AccessState().GetInt64(key)
	return value
}

func (o *ScState) GetObjectId(keyId int32, typeId int32) int32 {
	var factory func() WaspObject
	switch typeId {
	case OBJTYPE_BYTES_ARRAY, OBJTYPE_INT_ARRAY, OBJTYPE_STRING_ARRAY:
		//note that type of array elements can be found by decrementing typeId
		factory = func() WaspObject { return &ScStateArray{typeId: typeId - 1} }
	case OBJTYPE_MAP:
		factory = func() WaspObject { return &ScStateMap{} }
	default:
		o.error("GetObjectId: Invalid type")
		return 0
	}
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		keyId: {typeId, factory},
	})
}

func (o *ScState) GetString(keyId int32) string {
	if !o.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	key := o.vm.GetKey(keyId)
	value, _ := o.vm.ctx.AccessState().GetString(key)
	return value
}

func (o *ScState) SetBytes(keyId int32, value []byte) {
	if !o.valid(keyId, OBJTYPE_BYTES) {
		return
	}
	key := o.vm.GetKey(keyId)
	o.vm.ctx.AccessState().Set(key, value)
}

func (o *ScState) SetInt(keyId int32, value int64) {
	if !o.valid(keyId, OBJTYPE_INT) {
		return
	}
	key := o.vm.GetKey(keyId)
	o.vm.ctx.AccessState().SetInt64(key, value)
}

func (o *ScState) SetString(keyId int32, value string) {
	if !o.valid(keyId, OBJTYPE_STRING) {
		return
	}
	key := o.vm.GetKey(keyId)
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

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScStateArray struct {
	ArrayObject
	items  *datatypes.MustArray
	typeId int32
}

func (a *ScStateArray) InitVM(vm *wasmProcessor, keyId int32) {
	a.ArrayObject.InitVM(vm, 0)
	key := vm.GetKey(keyId)
	a.name = "state.array." + string(key)
	a.items = vm.ctx.AccessState().GetArray(key)
}

func (a *ScStateArray) Exists(keyId int32) bool {
	return keyId >= 0 && keyId < int32(a.items.Len())
}

func (a *ScStateArray) GetBytes(keyId int32) []byte {
	if !a.valid(keyId, OBJTYPE_BYTES) {
		return []byte(nil)
	}
	return a.items.GetAt(uint16(keyId))
}

func (a *ScStateArray) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.items.Len())
	}

	if !a.valid(keyId, OBJTYPE_INT) {
		return 0
	}
	value, _ := codec.DecodeInt64(a.items.GetAt(uint16(keyId)))
	return value
}

func (a *ScStateArray) GetString(keyId int32) string {
	if !a.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	return string(a.items.GetAt(uint16(keyId)))
}

func (a *ScStateArray) SetBytes(keyId int32, value []byte) {
	if !a.valid(keyId, OBJTYPE_BYTES) {
		return
	}
	a.items.SetAt(uint16(keyId), value)
}

func (a *ScStateArray) SetInt(keyId int32, value int64) {
	if keyId == KeyLength {
		a.items.Erase()
		return
	}
	if !a.valid(keyId, OBJTYPE_INT) {
		return
	}
	a.items.SetAt(uint16(keyId), util.Uint64To8Bytes(uint64(value)))
}

func (a *ScStateArray) SetString(keyId int32, value string) {
	if !a.valid(keyId, OBJTYPE_STRING) {
		return
	}
	a.items.SetAt(uint16(keyId), []byte(value))
}

func (a *ScStateArray) valid(keyId int32, typeId int32) bool {
	if a.typeId != typeId {
		a.error("valid: Invalid access")
		return false
	}
	max := int32(a.items.Len())
	if keyId == max {
		switch typeId {
		case OBJTYPE_BYTES:
			a.items.Push([]byte(nil))
		case OBJTYPE_INT:
			a.items.Push(util.Uint64To8Bytes(0))
		case OBJTYPE_STRING:
			a.items.Push([]byte(""))
		default:
			a.error("valid: Invalid type id")
			return false
		}
		return true
	}
	if keyId < 0 || keyId >= max {
		a.error("valid: Invalid index")
		return false
	}
	return true
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScStateMap struct {
	MapObject
	items *datatypes.MustMap
	types map[int32]int32
}

func (m *ScStateMap) InitVM(vm *wasmProcessor, keyId int32) {
	m.MapObject.InitVM(vm, 0)
	key := vm.GetKey(keyId)
	m.name = "state.map." + string(key)
	m.items = vm.ctx.AccessState().GetMap(key)
	m.types = make(map[int32]int32)
}

func (m *ScStateMap) Exists(keyId int32) bool {
	key := []byte(m.vm.GetKey(keyId))
	return m.items.HasAt(key)
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
	value, _ := codec.DecodeInt64(m.items.GetAt(key))
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
