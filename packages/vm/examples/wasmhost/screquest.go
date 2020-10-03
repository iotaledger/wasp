package wasmhost

import "github.com/iotaledger/wasp/packages/kv"

type ScRequest struct {
	MapObject
	balanceId int32
	colorsId  int32
	paramsId  int32
}

func NewScRequest(vm *wasmProcessor) HostObject {
	return &ScRequest{MapObject: MapObject{vm: vm, name: "Request"}}
}

func (o *ScRequest) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyTimestamp:
		return o.vm.ctx.GetTimestamp()
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScRequest) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyColors:
		return o.checkedObjectId(&o.colorsId, NewColorsArrayRequest, typeId, OBJTYPE_INT_ARRAY)
	case KeyBalance:
		return o.checkedObjectId(&o.balanceId, NewBalanceMapRequest, typeId, OBJTYPE_MAP)
	case KeyParams:
		return o.checkedObjectId(&o.paramsId, NewScRequestParams, typeId, OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScRequest) GetString(keyId int32) string {
	switch keyId {
	case KeyAddress:
		return o.vm.ctx.AccessRequest().Sender().String()
	case KeyHash:
		id := o.vm.ctx.AccessRequest().ID()
		return id.TransactionId().String()
	case KeyId:
		id := o.vm.ctx.AccessRequest().ID()
		return id.String()
	}
	return o.MapObject.GetString(keyId)
}

type ScRequestParams struct {
	MapObject
}

func NewScRequestParams(vm *wasmProcessor) HostObject {
	return &ScRequestParams{MapObject: MapObject{vm: vm, name: "Params"}}
}

func (o *ScRequestParams) GetInt(keyId int32) int64 {
	key := kv.Key(o.vm.GetKey(keyId))
	value, _, _ := o.vm.ctx.AccessRequest().Args().GetInt64(key)
	return value
}

func (o *ScRequestParams) GetString(keyId int32) string {
	key := kv.Key(o.vm.GetKey(keyId))
	value, _, _ := o.vm.ctx.AccessRequest().Args().GetString(key)
	return value
}
