package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

type ScLogs struct {
	MapObject
}

func (o *ScLogs) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		keyId: {OBJTYPE_MAP_ARRAY, func() WaspObject { return &ScLog{} }},
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScLog struct {
	ArrayObject
	lines      *datatypes.MustTimestampedLog
	logEntry   *ScLogEntry
	logEntryId int32
}

func (a *ScLog) InitVM(vm *wasmProcessor, keyId int32) {
	a.ModelObject.InitVM(vm, 0)
	key := vm.GetKey(keyId)
	a.name = "log." + string(key)
	a.lines = vm.ctx.AccessState().GetTimestampedLog(key)
	a.logEntry = &ScLogEntry{lines: a.lines}
	a.logEntryId = a.vm.TrackObject(a.logEntry)
}

func (a *ScLog) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.lines.Len())
	}
	return a.ModelObject.GetInt(keyId)
}

func (a *ScLog) GetObjectId(keyId int32, typeId int32) int32 {
	if typeId != OBJTYPE_MAP {
		a.error("GetObjectId: Invalid type")
		return 0
	}
	if keyId == int32(a.lines.Len()) {
		return a.logEntryId
	}
	return a.ArrayObject.GetObjectId(keyId, typeId)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScLogEntry struct {
	MapObject
	lines     *datatypes.MustTimestampedLog
	timestamp int64
}

func (o *ScLogEntry) Exists(keyId int32) bool {
	switch keyId {
	case KeyData:
	case KeyTimestamp:
	default:
		return false
	}
	return true
}

func (o *ScLogEntry) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyData:
		o.lines.Append(o.timestamp, value)
		return
	}
	o.error("SetBytes: Invalid key")
}

func (o *ScLogEntry) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyTimestamp:
		o.timestamp = value
		return
	}
	o.error("SetInt: Invalid key")
}
