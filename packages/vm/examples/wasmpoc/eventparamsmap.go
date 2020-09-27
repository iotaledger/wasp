package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasplib/host/interfaces"
)

type EventParamsMap struct {
	MapObject
	Params kv.Map
}

func NewEventParamsMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &EventParamsMap{MapObject: MapObject{vm: h, name: "EventParams"}, Params: kv.NewMap()}
}

func (o *EventParamsMap) GetBytes(keyId int32) []byte {
	value, _ := o.Params.Get(kv.Key(o.vm.GetKey(keyId)))
	return value
}

func (o *EventParamsMap) GetInt(keyId int32) int64 {
	value, ok, _ := o.Params.Codec().GetInt64(kv.Key(o.vm.GetKey(keyId)))
	if ok {
		return value
	}
	return o.MapObject.GetInt(keyId)
}

func (o *EventParamsMap) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *EventParamsMap) GetString(keyId int32) string {
	value, ok, _ := o.Params.Codec().GetString(kv.Key(o.vm.GetKey(keyId)))
	if ok {
		return value
	}
	return o.MapObject.GetString(keyId)
}

func (o *EventParamsMap) SetBytes(keyId int32, value []byte) {
	o.Params.Set(kv.Key(o.vm.GetKey(keyId)), value)
}

func (o *EventParamsMap) SetInt(keyId int32, value int64) {
	switch keyId {
	case interfaces.KeyLength:
		// clear params, tracker will still know about object
		// so maybe move it to an allocation pool for reuse
		o.Params = kv.NewMap()
	default:
		o.Params.Codec().SetInt64(kv.Key(o.vm.GetKey(keyId)), value)
	}
}

func (o *EventParamsMap) SetString(keyId int32, value string) {
	o.Params.Codec().SetString(kv.Key(o.vm.GetKey(keyId)), value)
}
