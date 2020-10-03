package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type ScEvents struct {
	ArrayObject
	events []int32
}

func NewScEvents(vm *wasmProcessor) HostObject {
	return &ScEvents{ArrayObject: ArrayObject{vm: vm, name: "Events"}}
}

func (a *ScEvents) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(len(a.events))
	}
	return a.ArrayObject.GetInt(keyId)
}

func (a *ScEvents) GetObjectId(keyId int32, typeId int32) int32 {
	return a.checkedObjectId(&a.events, keyId, NewScEvent, typeId, OBJTYPE_MAP)
}

func (a *ScEvents) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		for i := len(a.events) - 1; i >= 0; i-- {
			a.vm.SetInt(a.events[i], keyId, 0)
		}
		//todo move to pool for reuse of events?
		a.events = nil
		return
	default:
		a.ArrayObject.SetInt(keyId, value)
	}
}

type ScEvent struct {
	MapObject
	code     int64
	contract string
	delay    int64
	function string
	paramsId int32
}

func NewScEvent(vm *wasmProcessor) HostObject {
	return &ScEvent{MapObject: MapObject{vm: vm, name: "Event"}}
}

func (o *ScEvent) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyParams:
		return o.checkedObjectId(&o.paramsId, NewScEventParams, typeId, OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScEvent) Send() {
	o.vm.Trace("EVENT f'%s' c%d d%d a'%s'", o.function, o.code, o.delay, o.contract)
	if o.contract == "" {
		params := kv.NewMap()
		if o.paramsId != 0 {
			params = o.vm.FindObject(o.paramsId).(*ScEventParams).Params
			params.ForEach(func(key kv.Key, value []byte) bool {
				o.vm.Trace("  PARAM '%s'", key)
				return true
			})
		}
		if o.function != "" {
			params.Codec().SetString("fn", o.function)
			wasmPath, _, _ := o.vm.ctx.AccessRequest().Args().GetString("wasm")
			params.Codec().SetString("wasm", wasmPath)
		}
		if params.IsEmpty() {
			params = nil
		}
		o.vm.ctx.SendRequestToSelfWithDelay(sctransaction.RequestCode(uint16(o.code)), params, uint32(o.delay))
	}
}

func (o *ScEvent) SetInt(keyId int32, value int64) {
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
		o.Send()
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *ScEvent) SetString(keyId int32, value string) {
	switch keyId {
	case KeyContract:
		o.contract = value
	case KeyFunction:
		o.function = value
	default:
		o.MapObject.SetString(keyId, value)
	}
}

type ScEventParams struct {
	MapObject
	Params kv.Map
}

func NewScEventParams(vm *wasmProcessor) HostObject {
	return &ScEventParams{MapObject: MapObject{vm: vm, name: "EventParams"}, Params: kv.NewMap()}
}

func (o *ScEventParams) GetBytes(keyId int32) []byte {
	value, _ := o.Params.Get(kv.Key(o.vm.GetKey(keyId)))
	return value
}

func (o *ScEventParams) GetInt(keyId int32) int64 {
	value, ok, _ := o.Params.Codec().GetInt64(kv.Key(o.vm.GetKey(keyId)))
	if ok {
		return value
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScEventParams) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScEventParams) GetString(keyId int32) string {
	value, ok, _ := o.Params.Codec().GetString(kv.Key(o.vm.GetKey(keyId)))
	if ok {
		return value
	}
	return o.MapObject.GetString(keyId)
}

func (o *ScEventParams) SetBytes(keyId int32, value []byte) {
	o.Params.Set(kv.Key(o.vm.GetKey(keyId)), value)
}

func (o *ScEventParams) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		// clear params, tracker will still know about object
		// so maybe move it to an allocation pool for reuse
		o.Params = kv.NewMap()
	default:
		o.Params.Codec().SetInt64(kv.Key(o.vm.GetKey(keyId)), value)
	}
}

func (o *ScEventParams) SetString(keyId int32, value string) {
	o.Params.Codec().SetString(kv.Key(o.vm.GetKey(keyId)), value)
}
