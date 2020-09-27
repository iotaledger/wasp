package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasplib/host/interfaces"
)

type RequestParamsMap struct {
	MapObject
}

func NewParamsMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &RequestParamsMap{MapObject: MapObject{vm: h, name: "Params"}}
}

func (o *RequestParamsMap) GetInt(keyId int32) int64 {
	key := kv.Key(o.vm.GetKey(keyId))
	value, _, _ := o.vm.ctx.AccessRequest().Args().GetInt64(key)
	return value
}

func (o *RequestParamsMap) GetString(keyId int32) string {
	key := kv.Key(o.vm.GetKey(keyId))
	value, _, _ := o.vm.ctx.AccessRequest().Args().GetString(key)
	return value
}
