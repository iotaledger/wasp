package wasmhost

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ScCallInfo struct {
	MapObject
	contract string
	delay    int64
	function string
}

func (o *ScCallInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScCallInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyParams: func() WaspObject { return &ScCallParams{} },
	})
}

func (o *ScCallInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyContract:
		return OBJTYPE_STRING
	case KeyDelay:
		return OBJTYPE_INT
	case KeyFunction:
		return OBJTYPE_STRING
	case KeyParams:
		return OBJTYPE_MAP
	}
	return -1
}

func (o *ScCallInfo) Send() {
	o.vm.Trace("REQUEST c'%s' f'%s' d%d", o.contract, o.function, o.delay)
	if o.contract != "" {
		//TODO handle external contract
		o.vm.SetError("unknown contract")
		return
	}

	params := dict.New()
	paramsId, ok := o.objects[KeyParams]
	if ok {
		params = o.vm.FindObject(paramsId).(*ScCallParams).Params
		params.ForEach(func(key kv.Key, value []byte) bool {
			o.vm.Trace("  PARAM '%s'", key)
			return true
		})
	}
	if params.IsEmpty() {
		params = nil
	}
	reqCode := coretypes.Hn(o.function)
	if !o.vm.ctx.PostRequestToSelfWithDelay(reqCode, params, uint32(o.delay)) {
		o.vm.Trace("  FAILED to send")
	}
}

func (o *ScCallInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = ""
		o.delay = 0
		o.function = ""
	case KeyDelay:
		o.delay = value
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *ScCallInfo) SetString(keyId int32, value string) {
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

type ScCalls struct {
	ArrayObject
}

func (a *ScCalls) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		callInfo := &ScCallInfo{}
		callInfo.name = "call"
		return callInfo
	})
}

func (a *ScCalls) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return OBJTYPE_MAP
	}
	return -1
}

func (a *ScCalls) Send() {
	for i := 0; i < len(a.objects); i++ {
		request := a.vm.FindObject(a.objects[i]).(*ScCallInfo)
		request.Send()
	}
}

func (a *ScCalls) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		a.objects = nil
		return
	default:
		a.ArrayObject.SetInt(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCallParams struct {
	MapObject
	Params dict.Dict
}

func (o *ScCallParams) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.Params = dict.New()
}

func (o *ScCallParams) Exists(keyId int32) bool {
	key := o.vm.GetKey(keyId)
	exists, _ := o.Params.Has(key)
	return exists
}

func (o *ScCallParams) GetBytes(keyId int32) []byte {
	key := o.vm.GetKey(keyId)
	value, _ := o.Params.Get(key)
	return value
}

func (o *ScCallParams) GetInt(keyId int32) int64 {
	key := o.vm.GetKey(keyId)
	bytes, err := o.Params.Get(key)
	if err == nil {
		value, err := codec.DecodeInt64(bytes)
		if err == nil {
			return value
		}
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScCallParams) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScCallParams) GetString(keyId int32) string {
	key := o.vm.GetKey(keyId)
	bytes, err := o.Params.Get(key)
	if err == nil {
		return codec.DecodeString(bytes)
	}
	return o.MapObject.GetString(keyId)
}

//TODO keep track of field types
func (o *ScCallParams) GetTypeId(keyId int32) int32 {
	return o.MapObject.GetTypeId(keyId)
}

func (o *ScCallParams) SetBytes(keyId int32, value []byte) {
	key := o.vm.GetKey(keyId)
	o.Params.Set(key, value)
}

func (o *ScCallParams) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Params = dict.New()
	default:
		key := o.vm.GetKey(keyId)
		o.Params.Set(key, codec.EncodeInt64(value))
	}
}

func (o *ScCallParams) SetString(keyId int32, value string) {
	key := o.vm.GetKey(keyId)
	o.Params.Set(key, codec.EncodeString(value))
}
