package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type ParamsObject struct {
	vm *wasmVMPocProcessor
}

func NewParamsObject(h *wasmVMPocProcessor) *ParamsObject {
	return &ParamsObject{vm: h}
}

func (o *ParamsObject) GetInt(keyId int32) int64 {
	key := kv.Key(o.vm.GetKey(keyId))
	value, _, _ := o.vm.ctx.AccessRequest().Args().GetInt64(key)
	return value
}

func (o *ParamsObject) GetObjectId(keyId int32, typeId int32) int32 {
	panic("implement Params.GetObjectId")
}

func (o *ParamsObject) GetString(keyId int32) string {
	key := kv.Key(o.vm.GetKey(keyId))
	value, _, _ := o.vm.ctx.AccessRequest().Args().GetString(key)
	return value
}

func (o *ParamsObject) SetInt(keyId int32, value int64) {
	o.vm.SetError("Readonly Params.SetInt")
}

func (o *ParamsObject) SetString(keyId int32, value string) {
	o.vm.SetError("Readonly Params.SetString")
}
