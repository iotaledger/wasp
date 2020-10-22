package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type ScPostedRequest struct {
	MapObject
	code     int64
	contract string
	delay    int64
}

func (o *ScPostedRequest) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyParams: {OBJTYPE_MAP, func() WaspObject { return &ScPostParams{} }},
	})
}

func (o *ScPostedRequest) Send() {
	o.vm.Trace("REQUEST f'%s' c%d d%d a'%s'", o.vm.codeToFunc[int32(o.code)], o.code, o.delay, o.contract)
	if o.contract == o.vm.ctx.GetSCAddress().String() {
		params := kv.NewMap()
		paramsId, ok := o.objects[KeyParams]
		if ok {
			params = o.vm.FindObject(paramsId).(*ScPostParams).Params
			params.ForEach(func(key kv.Key, value []byte) bool {
				o.vm.Trace("  PARAM '%s'", key)
				return true
			})
		}
		if params.IsEmpty() {
			params = nil
		}
		o.vm.ctx.SendRequestToSelfWithDelay(sctransaction.RequestCode(o.code), params, uint32(o.delay))
	}
	//TODO handle external contract
}

func (o *ScPostedRequest) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = ""
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

func (o *ScPostedRequest) SetString(keyId int32, value string) {
	switch keyId {
	case KeyContract:
		o.contract = value
	case KeyFunction:
		code, ok := o.vm.funcToCode[value]
		if !ok {
			o.error("SetString: invalid function: %s", value)
			return
		}
		o.code = int64(code)
	default:
		o.MapObject.SetString(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScPostedRequests struct {
	ArrayObject
}

func (a *ScPostedRequests) GetObjectId(keyId int32, typeId int32) int32 {
	return a.GetArrayObjectId(keyId, typeId, func() WaspObject {
		postedRequest := &ScPostedRequest{}
		postedRequest.name = "postedRequest"
		return postedRequest
	})
}

func (a *ScPostedRequests) Send() {
	for i := 0; i < len(a.objects); i++ {
		request := a.vm.FindObject(a.objects[i]).(*ScPostedRequest)
		request.Send()
	}
}

func (a *ScPostedRequests) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		a.objects = nil
		return
	default:
		a.ArrayObject.SetInt(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScPostParams struct {
	MapObject
	Params kv.Map
}

func (o *ScPostParams) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.Params = kv.NewMap()
}

func (o *ScPostParams) GetBytes(keyId int32) []byte {
	value, _ := o.Params.Get(o.vm.GetKey(keyId))
	return value
}

func (o *ScPostParams) GetInt(keyId int32) int64 {
	value, ok, _ := o.Params.Codec().GetInt64(o.vm.GetKey(keyId))
	if ok {
		return value
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScPostParams) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScPostParams) GetString(keyId int32) string {
	value, ok, _ := o.Params.Codec().GetString(o.vm.GetKey(keyId))
	if ok {
		return value
	}
	return o.MapObject.GetString(keyId)
}

func (o *ScPostParams) SetBytes(keyId int32, value []byte) {
	o.Params.Set(o.vm.GetKey(keyId), value)
}

func (o *ScPostParams) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Params = kv.NewMap()
	default:
		o.Params.Codec().SetInt64(o.vm.GetKey(keyId), value)
	}
}

func (o *ScPostParams) SetString(keyId int32, value string) {
	o.Params.Codec().SetString(o.vm.GetKey(keyId), value)
}
