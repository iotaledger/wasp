package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type EventMap struct {
	MapObject
	code     int64
	contract string
	delay    int64
	function string
	paramsId int32
}

func NewEventMap(vm *wasmVMPocProcessor) HostObject {
	return &EventMap{MapObject: MapObject{vm: vm, name: "Event"}}
}

func (o *EventMap) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyParams:
		return o.checkedObjectId(&o.paramsId, NewEventParamsMap, typeId, OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *EventMap) Send() {
	o.vm.Trace("EVENT f'%s' c%d d%d a'%s'", o.function, o.code, o.delay, o.contract)
	if o.contract == "" {
		params := kv.NewMap()
		if o.paramsId != 0 {
			params = o.vm.FindObject(o.paramsId).(*EventParamsMap).Params
			params.ForEach(func(key kv.Key, value []byte) bool {
				o.vm.Trace("  PARAM '%s'", key)
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
	case KeyLength:
		// clear request, tracker will still know about object
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
