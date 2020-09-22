package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces/objtype"
)

type RequestMap struct {
	MapObject
	balanceId int32
	colorsId  int32
	paramsId  int32
}

func NewRequestMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &RequestMap{MapObject: MapObject{vm: h, name: "Request"}}
}

func (o *RequestMap) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyTimestamp:
		return o.vm.ctx.GetTimestamp()
	}
	return o.MapObject.GetInt(keyId)
}

func (o *RequestMap) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyColors:
		return o.checkedObjectId(&o.colorsId, NewColorsArrayRequest, typeId, objtype.OBJTYPE_INT_ARRAY)
	case KeyBalance:
		return o.checkedObjectId(&o.balanceId, NewBalanceMapRequest, typeId, objtype.OBJTYPE_MAP)
	case KeyParams:
		return o.checkedObjectId(&o.paramsId, NewParamsMap, typeId, objtype.OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *RequestMap) GetString(keyId int32) string {
	switch keyId {
	case KeyAddress:
		return o.vm.ctx.AccessRequest().Sender().String()
	case KeyHash:
		id := o.vm.ctx.AccessRequest().ID()
		return id.String()
	}
	return o.MapObject.GetString(keyId)
}
