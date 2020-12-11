// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/util"
)

type ScState struct {
	ScMutableDict
}

func (o *ScState) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.ScMutableDict.InitObj(id, keyId, owner)
	o.Dict = o.vm.State()
}

func (o *ScState) GetObjectId(keyId int32, typeId int32) int32 {
	o.validate(keyId, typeId)
	var factory ObjFactory
	switch typeId {
	case OBJTYPE_BYTES_ARRAY, OBJTYPE_INT_ARRAY, OBJTYPE_STRING_ARRAY:
		//note that type of array elements can be found by decrementing typeId
		factory = func() WaspObject { return &ScStateArray{typeId: typeId - 1} }
	case OBJTYPE_MAP:
		factory = func() WaspObject { return &ScStateMap{} }
	default:
		o.Panic("GetObjectId: Invalid type")
	}
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		keyId: factory,
	})
}

func (o *ScState) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		//TODO how to clear state?
		o.Panic("SetInt: Clear state")
	default:
		o.ScMutableDict.SetInt(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScStateArray struct {
	ArrayObject
	items  *datatypes.MustArray
	typeId int32
}

func (a *ScStateArray) InitObj(id int32, keyId int32, owner *ModelObject) {
	a.ArrayObject.InitObj(id, keyId, owner)
	key := a.vm.GetKey(keyId)
	a.items = datatypes.NewMustArray(a.vm.State(), string(key))
}

func (a *ScStateArray) Exists(keyId int32) bool {
	return uint32(keyId) <= uint32(a.items.Len())
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
	value, _, err := codec.DecodeInt64(a.items.GetAt(uint16(keyId)))
	if err != nil {
		panic(err)
	}
	return value
}

func (a *ScStateArray) GetString(keyId int32) string {
	if !a.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	return string(a.items.GetAt(uint16(keyId)))
}

func (a *ScStateArray) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return a.typeId
	}
	return -1
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
		a.Panic("valid: Invalid access")
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
			a.Panic("valid: Invalid type id")
		}
		return true
	}
	if keyId < 0 || keyId >= max {
		a.Panic("valid: Invalid index")
	}
	return true
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScStateMap struct {
	MapObject
	items *datatypes.MustMap
	types map[int32]int32
}

func (m *ScStateMap) InitObj(id int32, keyId int32, owner *ModelObject) {
	m.MapObject.InitObj(id, keyId, owner)
	key := m.vm.GetKey(keyId)
	m.items = datatypes.NewMustMap(m.vm.State(), string(key))
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
	value, _, err := codec.DecodeInt64(m.items.GetAt(key))
	if err != nil {
		panic(err)
	}
	return value
}

func (m *ScStateMap) GetObjectId(keyId int32, typeId int32) int32 {
	m.Panic("GetObjectId: Invalid access")
	return 0
}

func (m *ScStateMap) GetString(keyId int32) string {
	if !m.valid(keyId, OBJTYPE_STRING) {
		return ""
	}
	key := []byte(m.vm.GetKey(keyId))
	return string(m.items.GetAt(key))
}

func (m *ScStateMap) GetTypeId(keyId int32) int32 {
	typeId, ok := m.types[keyId]
	if ok {
		return typeId
	}
	return -1
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
		m.Panic("SetInt: Invalid clear")
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
		m.Panic("valid: Invalid access")
	}
	return true
}
