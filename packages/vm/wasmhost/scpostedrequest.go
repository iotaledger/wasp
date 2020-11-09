package wasmhost

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ScPostedRequest struct {
	MapObject
	code     int64
	contract []byte
	delay    int64
}

func (o *ScPostedRequest) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyParams: {OBJTYPE_MAP, func() WaspObject { return &ScPostParams{} }},
	})
}

func (o *ScPostedRequest) Send() {
	function := o.vm.codeToFunc[int32(o.code)]
	o.vm.Trace("REQUEST f'%s' c%d d%d a'%s'", function, o.code, o.delay, o.contract)
	contractID := o.vm.ctx.GetContractID()
	if bytes.Equal(o.contract, contractID[:coretypes.ChainIDLength]) {
		params := dict.NewDict()
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
		o.vm.ctx.SendRequestToSelfWithDelay(coretypes.Hname(o.code), params, uint32(o.delay))
	}
	//TODO handle external contract
}

func (o *ScPostedRequest) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyContract:
		o.contract = value
	default:
		o.MapObject.SetBytes(keyId, value)
	}
}

func (o *ScPostedRequest) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = nil
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
	Params dict.Dict
}

func (o *ScPostParams) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.Params = dict.NewDict()
}

func (o *ScPostParams) GetBytes(keyId int32) []byte {
	value, _ := o.Params.Get(o.vm.GetKey(keyId))
	return value
}

func (o *ScPostParams) GetInt(keyId int32) int64 {
	bytes, err := o.Params.Get(o.vm.GetKey(keyId))
	if err == nil {
		value, err := codec.DecodeInt64(bytes)
		if err == nil {
			return value
		}
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScPostParams) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScPostParams) GetString(keyId int32) string {
	bytes, err := o.Params.Get(o.vm.GetKey(keyId))
	if err == nil {
		return codec.DecodeString(bytes)
	}
	return o.MapObject.GetString(keyId)
}

func (o *ScPostParams) SetBytes(keyId int32, value []byte) {
	o.Params.Set(o.vm.GetKey(keyId), value)
}

func (o *ScPostParams) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Params = dict.NewDict()
	default:
		o.Params.Set(o.vm.GetKey(keyId), codec.EncodeInt64(value))
	}
}

func (o *ScPostParams) SetString(keyId int32, value string) {
	o.Params.Set(o.vm.GetKey(keyId), codec.EncodeString(value))
}
