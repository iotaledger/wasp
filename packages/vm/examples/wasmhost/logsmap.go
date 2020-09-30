package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type LogsMap struct {
	MapObject
	logs map[int32]int32
}

func NewLogsMap(vm *wasmVMPocProcessor) HostObject {
	return &LogsMap{MapObject: MapObject{vm: vm, name: "Logs"}, logs: make(map[int32]int32)}
}

func (o *LogsMap) GetObjectId(keyId int32, typeId int32) int32 {
	if typeId != OBJTYPE_MAP {
		o.vm.SetError("Invalid type id")
		return 0
	}
	objId, ok := o.logs[keyId]
	if !ok {
		key := kv.Key(o.vm.GetKey(keyId))
		a := o.vm.ctx.AccessState().GetTimestampedLog(key)
		o.logs[keyId] = o.vm.AddObject(NewLogMap(o.vm, a))
	}
	return objId
}
