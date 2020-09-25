package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces/objtype"
)

type EventMap struct {
	MapObject
	code     int64
	contract string
	delay    int64
	function string
	paramsId int32
}

func NewEventMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &EventMap{MapObject: MapObject{vm: h, name: "Event"}}
}

func (o *EventMap) GetInt(keyId int32) int64 {
	switch keyId {
	//case interfaces.KeyCode:
	//	return o.code
	//case interfaces.KeyDelay:
	//	return o.delay
	}
	return o.MapObject.GetInt(keyId)
}

func (o *EventMap) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyParams:
		return o.checkedObjectId(&o.paramsId, NewEventParamsMap, typeId, objtype.OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *EventMap) GetString(keyId int32) string {
	switch keyId {
	//case interfaces.KeyContract:
	//	return o.contract
	//case interfaces.KeyFunction:
	//	return o.function
	}
	return o.MapObject.GetString(keyId)
}

func (o *EventMap) Send() {
	o.vm.Logf("REQ SEND c%d d%d a'%s'", o.code, o.delay, o.contract)
	if o.contract == "" {
		var params kv.Map = kv.NewMap()
		if o.paramsId != 0 {
			params = o.vm.GetObject(o.paramsId).(*EventParamsMap).Params
			params.ForEach(func(key kv.Key, value []byte) bool {
				o.vm.Logf("  PARAM '%s'", key)
				return true
			})
		}
		if o.function != "" {
			params.Codec().SetString("fn", o.function)
		}
		if params.IsEmpty() {
			params = nil
		}
		o.vm.ctx.SendRequestToSelfWithDelay(sctransaction.RequestCode(uint16(o.code)), params, uint32(o.delay))
	}
}

func (o *EventMap) SetInt(keyId int32, value int64) {
	switch keyId {
	case interfaces.KeyLength:
		// clear request, tracker will still know about it
		// so maybe move it to an allocation pool for reuse
		o.contract = ""
		o.function = ""
		o.code = 0
		o.delay = 0
	case KeyCode:
		o.code = value
	case KeyDelay:
		o.delay = value
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *EventMap) SetString(keyId int32, value string) {
	switch keyId {
	case KeyContract:
		o.contract = value
	case KeyFunction:
		o.function = value
	default:
		o.MapObject.SetString(keyId, value)
	}
}
