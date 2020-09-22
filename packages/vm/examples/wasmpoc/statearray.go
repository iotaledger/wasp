package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces/objtype"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

type StateArray struct {
	ArrayObject
	items  *kv.MustArray
	typeId int32
}

func NewStateArray(h *wasmVMPocProcessor, items *kv.MustArray, typeId int32) interfaces.HostObject {
	return &StateArray{ArrayObject: ArrayObject{vm: h, name: "StateArray"}, items: items, typeId: typeId}
}

func (a *StateArray) GetInt(keyId int32) int64 {
	switch keyId {
	case interfaces.KeyLength:
		return int64(a.GetLength())
	}

	if !a.valid(keyId, objtype.OBJTYPE_INT) {
		return 0
	}
	value, _ := kv.DecodeInt64(a.items.GetAt(uint16(keyId)))
	return value
}

func (a *StateArray) GetLength() int32 {
	return int32(a.items.Len())
}

func (a *StateArray) GetString(keyId int32) string {
	if !a.valid(keyId, objtype.OBJTYPE_STRING) {
		return ""
	}
	return string(a.items.GetAt(uint16(keyId)))
}

func (a *StateArray) SetInt(keyId int32, value int64) {
	if keyId == interfaces.KeyLength {
		a.items.Erase()
		return
	}
	if !a.valid(keyId, objtype.OBJTYPE_INT) {
		return
	}
	a.items.SetAt(uint16(keyId), util.Uint64To8Bytes(uint64(value)))
}

func (a *StateArray) SetString(keyId int32, value string) {
	if !a.valid(keyId, objtype.OBJTYPE_STRING) {
		return
	}
	a.items.SetAt(uint16(keyId), []byte(value))
}

func (a *StateArray) valid(keyId int32, typeId int32) bool {
	if a.typeId != typeId {
		a.error("valid: Invalid access")
		return false
	}
	max := a.GetLength()
	if keyId == max {
		switch typeId {
		case objtype.OBJTYPE_INT:
			a.items.Push(util.Uint64To8Bytes(0))
		case objtype.OBJTYPE_STRING:
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
