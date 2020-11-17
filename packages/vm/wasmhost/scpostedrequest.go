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
	contract coretypes.AgentID
	delay    int64
	funcCode uint32
}

func (o *ScPostedRequest) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScPostedRequest) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyParams: func() WaspObject { return &ScPostParams{} },
	})
}

func (o *ScPostedRequest) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyContract:
		return OBJTYPE_BYTES
	case KeyDelay:
		return OBJTYPE_INT
	case KeyFunction:
		return OBJTYPE_STRING
	case KeyParams:
		return OBJTYPE_MAP
	}
	return -1
}

func (o *ScPostedRequest) Send() {
	function := o.vm.codeToFunc[o.funcCode]
	o.vm.Trace("REQUEST f'%s' c%d d%d a'%s'", function, o.funcCode, o.delay, o.contract.String())
	contract := o.vm.ctx.MyContractID()
	if !bytes.Equal(o.contract[:], contract[:]) {
		//TODO handle external contract
		o.vm.Trace("Unknown contract id")
		return
	}

	params := dict.New()
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
	reqCode := coretypes.Hname(o.funcCode)
	if !o.vm.ctx.PostRequestToSelfWithDelay(reqCode, params, uint32(o.delay)) {
		o.vm.Trace("  FAILED to send")
	}
}

func (o *ScPostedRequest) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyContract:
		o.contract,_ = coretypes.NewAgentIDFromBytes(value)
	default:
		o.MapObject.SetBytes(keyId, value)
	}
}

func (o *ScPostedRequest) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = coretypes.AgentID{}
		o.delay = 0
		o.funcCode = 0
	case KeyDelay:
		o.delay = value
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *ScPostedRequest) SetString(keyId int32, value string) {
	switch keyId {
	case KeyFunction:
		funcCode, ok := o.vm.funcToCode[value]
		if !ok {
			o.Error("SetString: invalid function: %s", value)
			return
		}
		o.funcCode = funcCode
	default:
		o.MapObject.SetString(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScPostedRequests struct {
	ArrayObject
}

func (a *ScPostedRequests) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		postedRequest := &ScPostedRequest{}
		postedRequest.name = "postedRequest"
		return postedRequest
	})
}

func (a *ScPostedRequests) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return OBJTYPE_MAP
	}
	return -1
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
	o.Params = dict.New()
}

func (o *ScPostParams) Exists(keyId int32) bool {
	key := o.vm.GetKey(keyId)
	exists, _ := o.Params.Has(key)
	return exists
}

func (o *ScPostParams) GetBytes(keyId int32) []byte {
	key := o.vm.GetKey(keyId)
	value, _ := o.Params.Get(key)
	return value
}

func (o *ScPostParams) GetInt(keyId int32) int64 {
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

func (o *ScPostParams) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScPostParams) GetString(keyId int32) string {
	key := o.vm.GetKey(keyId)
	bytes, err := o.Params.Get(key)
	if err == nil {
		return codec.DecodeString(bytes)
	}
	return o.MapObject.GetString(keyId)
}

//TODO keep track of field types
func (o *ScPostParams) GetTypeId(keyId int32) int32 {
	return o.MapObject.GetTypeId(keyId)
}

func (o *ScPostParams) SetBytes(keyId int32, value []byte) {
	key := o.vm.GetKey(keyId)
	o.Params.Set(key, value)
}

func (o *ScPostParams) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Params = dict.New()
	default:
		key := o.vm.GetKey(keyId)
		o.Params.Set(key, codec.EncodeInt64(value))
	}
}

func (o *ScPostParams) SetString(keyId int32, value string) {
	key := o.vm.GetKey(keyId)
	o.Params.Set(key, codec.EncodeString(value))
}
