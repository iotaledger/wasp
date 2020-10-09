package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type ScEvent struct {
	MapObject
	code     int64
	contract string
	delay    int64
	function string
}

func (o *ScEvent) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyParams: {OBJTYPE_MAP, func() WaspObject { return &ScEventParams{} }},
	})
}

func (o *ScEvent) Send() {
	o.vm.Trace("EVENT f'%s' c%d d%d a'%s'", o.function, o.code, o.delay, o.contract)
	if o.contract == "" {
		params := kv.NewMap()
		paramsId, ok := o.objects[KeyParams]
		if ok {
			params = o.vm.FindObject(paramsId).(*ScEventParams).Params
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

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScEvents struct {
	ArrayObject
}

func (a *ScEvents) Clear() {
	for i := len(a.objects) - 1; i >= 0; i-- {
		a.vm.SetInt(a.objects[i], KeyLength, 0)
	}
	//TODO move to pool for reuse of events?
	a.objects = nil
}

func (a *ScEvents) GetObjectId(keyId int32, typeId int32) int32 {
	return a.GetArrayObjectId(keyId, typeId, func() WaspObject {
		event := &ScEvent{}
		event.name = "event"
		return event
	})
}

func (a *ScEvents) Send() {
	for i := 0; i < len(a.objects); i++ {
		request := a.vm.FindObject(a.objects[i]).(*ScEvent)
		request.Send()
	}
	a.Clear()
}

func (a *ScEvents) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		a.Clear()
		return
	default:
		a.ArrayObject.SetInt(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScEventParams struct {
	MapObject
	Params kv.Map
}

func (o *ScEventParams) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.Params = kv.NewMap()
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
		//TODO clear kv map?
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
