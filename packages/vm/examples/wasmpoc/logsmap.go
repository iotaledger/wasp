package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces/objtype"
)

type LogsMap struct {
	MapObject
	logs map[int32]int32
}

func NewLogsMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &LogsMap{MapObject: MapObject{vm: h, name: "Logs"}, logs: make(map[int32]int32)}
}

func (o *LogsMap) GetObjectId(keyId int32, typeId int32) int32 {
	if typeId != objtype.OBJTYPE_MAP {
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
