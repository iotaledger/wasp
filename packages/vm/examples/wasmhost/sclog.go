package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type LogsMap struct {
	MapObject
	logs map[int32]int32
}

func (o *LogsMap) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.logs= make(map[int32]int32)
}

func (o *LogsMap) GetObjectId(keyId int32, typeId int32) int32 {
	if typeId != OBJTYPE_MAP {
		o.error("GetObjectId: Invalid type id")
		return 0
	}
	objId, ok := o.logs[keyId]
	if !ok {
		key := kv.Key(o.vm.GetKey(keyId))
		a := o.vm.ctx.AccessState().GetTimestampedLog(key)
		o.logs[keyId] = o.vm.TrackObject(NewLogMap(o.vm, a))
	}
	return objId
}

type LogMap struct {
	MapObject
	lines     *kv.MustTimestampedLog
	timestamp int64
}

func NewLogMap(vm *wasmProcessor, a *kv.MustTimestampedLog) HostObject {
	return &LogMap{MapObject: MapObject{vm: vm, name: "LogMap"}, lines: a}
}

func (a *LogMap) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.lines.Len())
	}
	return a.MapObject.GetInt(keyId)
}

func (a *LogMap) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyData:
		a.lines.Append(a.timestamp, value)
		return
	}
	a.error("SetBytes: Invalid key")
}

func (a *LogMap) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyTimestamp:
		a.timestamp = value
		return
	}
	a.error("SetInt: Invalid key")
}
