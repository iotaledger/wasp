package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type ScLog struct {
	ModelObject
	lines     *kv.MustTimestampedLog
	timestamp int64
}

func (a *ScLog) InitVM(vm *wasmProcessor, keyId int32) {
	a.ModelObject.InitVM(vm, 0)
	key := vm.GetKey(keyId)
	a.name = "log." + key
	a.lines = vm.ctx.AccessState().GetTimestampedLog(kv.Key(key))
}

func (a *ScLog) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.lines.Len())
	}
	return a.ModelObject.GetInt(keyId)
}

func (a *ScLog) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyData:
		a.lines.Append(a.timestamp, value)
		return
	}
	a.error("SetBytes: Invalid key")
}

func (a *ScLog) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyTimestamp:
		a.timestamp = value
		return
	}
	a.error("SetInt: Invalid key")
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScLogs struct {
	MapObject
}

func (o *ScLogs) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		keyId: {OBJTYPE_MAP, func() WaspObject { return &ScLog{} }},
	})
}
