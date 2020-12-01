package wasmhost

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type ScCallInfo struct {
	MapObject
	contract string
	function string
}

func (o *ScCallInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScCallInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyParams:    func() WaspObject { return &ScCallParams{} },
		KeyResults:   func() WaspObject { return &ScCallResults{} },
		KeyTransfers: func() WaspObject { return &ScCallTransfers{} },
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
	case KeyResults:
		return OBJTYPE_MAP
	case KeyTransfers:
		return OBJTYPE_MAP
	}
	return -1
}

func (o *ScCallInfo) Invoke() {
	o.vm.Trace("CALL c'%s' f'%s'", o.contract, o.function)
	contractCode := o.vm.MyContractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	paramsId := o.GetObjectId(KeyParams, OBJTYPE_MAP)
	params := o.vm.FindObject(paramsId).(*ScCallParams).Params
	params.ForEach(func(key kv.Key, value []byte) bool {
		o.vm.Trace("  PARAM '%s'", key)
		return true
	})
	transfersId := o.GetObjectId(KeyTransfers, OBJTYPE_MAP)
	transfers := o.vm.FindObject(transfersId).(*ScCallTransfers).Transfers
	balances := cbalances.NewFromMap(transfers)
	var err error
	var results codec.ImmutableCodec
	if o.vm.ctx != nil {
		results, err = o.vm.ctx.Call(contractCode, functionCode, codec.NewCodec(params), balances)
	} else {
		results, err = o.vm.ctxView.Call(contractCode, functionCode, codec.NewCodec(params))
	}
	if err != nil {
		o.Error("failed to invoke call: %v", err)
	}
	resultsId := o.GetObjectId(KeyResults, OBJTYPE_MAP)
	o.vm.FindObject(resultsId).(*ScCallResults).Results = results
}

func (o *ScCallInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = ""
		o.function = ""
	case KeyDelay:
		if value != -1 {
			o.Error("Unexpected value for delay: %d", value)
		}
		o.Invoke()
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

type ScPostInfo struct {
	MapObject
	chainId  *coretypes.ChainID
	contract string
	delay    uint32
	function string
}

func (o *ScPostInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScPostInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyParams:    func() WaspObject { return &ScCallParams{} },
		KeyTransfers: func() WaspObject { return &ScCallTransfers{} },
	})
}

func (o *ScPostInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyChain:
		return OBJTYPE_BYTES
	case KeyContract:
		return OBJTYPE_STRING
	case KeyDelay:
		return OBJTYPE_INT
	case KeyFunction:
		return OBJTYPE_STRING
	case KeyParams:
		return OBJTYPE_MAP
	case KeyTransfers:
		return OBJTYPE_MAP
	}
	return -1
}

func (o *ScPostInfo) Invoke() {
	o.vm.Trace("POST c'%s' f'%s' d%d", o.contract, o.function, o.delay)
	chainId := o.vm.ctx.ChainID()
	if o.chainId != nil {
		chainId = *o.chainId
	}
	contractCode := o.vm.MyContractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	params := dict.New()
	paramsId, ok := o.objects[KeyParams]
	if ok {
		params = o.vm.FindObject(paramsId).(*ScCallParams).Params
		params.ForEach(func(key kv.Key, value []byte) bool {
			o.vm.Trace("  PARAM '%s'", key)
			return true
		})
	}
	transfersId := o.GetObjectId(KeyTransfers, OBJTYPE_MAP)
	transfers := o.vm.FindObject(transfersId).(*ScCallTransfers).Transfers
	balances := cbalances.NewFromMap(transfers)
	if !o.vm.ctx.PostRequest(vmtypes.NewRequestParams{
		TargetContractID: coretypes.NewContractID(chainId, contractCode),
		EntryPoint:       functionCode,
		Params:           params,
		Timelock:         util.NanoSecToUnixSec(o.vm.ctx.GetTimestamp()) + o.delay,
		Transfer:         balances,
	}) {
		o.Error("failed to invoke post")
	}
}

func (o *ScPostInfo) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyChain:
		chainId, err := coretypes.NewChainIDFromBytes(value)
		if err != nil {
			o.Error(err.Error())
			return
		}
		o.chainId = &chainId
	default:
		o.MapObject.SetBytes(keyId, value)
	}
}

func (o *ScPostInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.chainId = nil
		o.contract = ""
		o.delay = 0
		o.function = ""
	case KeyDelay:
		if value < 0 {
			o.Error("Unexpected value for delay: %d", value)
		}
		o.delay = uint32(value)
		o.Invoke()
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *ScPostInfo) SetString(keyId int32, value string) {
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

type ScViewInfo struct {
	MapObject
	contract string
	function string
}

func (o *ScViewInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScViewInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyParams:  func() WaspObject { return &ScCallParams{} },
		KeyResults: func() WaspObject { return &ScCallResults{} },
	})
}

func (o *ScViewInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyContract:
		return OBJTYPE_STRING
	case KeyDelay:
		return OBJTYPE_INT
	case KeyFunction:
		return OBJTYPE_STRING
	case KeyParams:
		return OBJTYPE_MAP
	case KeyResults:
		return OBJTYPE_MAP
	}
	return -1
}

func (o *ScViewInfo) Invoke() {
	o.vm.Trace("VIEW c'%s' f'%s'", o.contract, o.function)
	contractCode := o.vm.MyContractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	params := dict.New()
	paramsId, ok := o.objects[KeyParams]
	if ok {
		params = o.vm.FindObject(paramsId).(*ScCallParams).Params
		params.ForEach(func(key kv.Key, value []byte) bool {
			o.vm.Trace("  PARAM '%s'", key)
			return true
		})
	}
	var err error
	var results codec.ImmutableCodec
	if o.vm.ctx != nil {
		results, err = o.vm.ctx.Call(contractCode, functionCode, codec.NewCodec(params), nil)
	} else {
		results, err = o.vm.ctxView.Call(contractCode, functionCode, codec.NewCodec(params))
	}
	if err != nil {
		o.Error("failed to invoke view: %v", err)
	}
	resultsId := o.GetObjectId(KeyResults, OBJTYPE_MAP)
	o.vm.FindObject(resultsId).(*ScCallResults).Results = results
}

func (o *ScViewInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = ""
		o.function = ""
	case KeyDelay:
		if value != -2 {
			o.Error("Unexpected value for delay: %d", value)
		}
		o.Invoke()
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *ScViewInfo) SetString(keyId int32, value string) {
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

type ScPosts struct {
	ArrayObject
}

func (a *ScPosts) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		postInfo := &ScPostInfo{}
		postInfo.name = "post"
		return postInfo
	})
}

func (a *ScPosts) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return OBJTYPE_MAP
	}
	return -1
}

func (a *ScPosts) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		a.objects = nil
		return
	default:
		a.ArrayObject.SetInt(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScViews struct {
	ArrayObject
}

func (a *ScViews) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		viewInfo := &ScViewInfo{}
		viewInfo.name = "view"
		return viewInfo
	})
}

func (a *ScViews) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return OBJTYPE_MAP
	}
	return -1
}

func (a *ScViews) SetInt(keyId int32, value int64) {
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

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCallResults struct {
	MapObject
	Results codec.ImmutableCodec
}

func (o *ScCallResults) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.Results = codec.NewCodec(dict.New())
}

func (o *ScCallResults) Exists(keyId int32) bool {
	key := o.vm.GetKey(keyId)
	exists, _ := o.Results.Has(key)
	return exists
}

func (o *ScCallResults) GetBytes(keyId int32) []byte {
	key := o.vm.GetKey(keyId)
	value, _ := o.Results.Get(key)
	return value
}

func (o *ScCallResults) GetInt(keyId int32) int64 {
	key := o.vm.GetKey(keyId)
	bytes, err := o.Results.Get(key)
	if err == nil {
		value, err := codec.DecodeInt64(bytes)
		if err == nil {
			return value
		}
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScCallResults) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScCallResults) GetString(keyId int32) string {
	key := o.vm.GetKey(keyId)
	bytes, err := o.Results.Get(key)
	if err == nil {
		return codec.DecodeString(bytes)
	}
	return o.MapObject.GetString(keyId)
}

//TODO keep track of field types
func (o *ScCallResults) GetTypeId(keyId int32) int32 {
	return o.MapObject.GetTypeId(keyId)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCallTransfers struct {
	MapObject
	Transfers map[balance.Color]int64
}

func (o *ScCallTransfers) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.Transfers = make(map[balance.Color]int64)
}

func (o *ScCallTransfers) Exists(keyId int32) bool {
	var color balance.Color = [32]byte{}
	copy(color[:], o.vm.getKeyFromId(keyId))
	return o.Transfers[color] != 0
}

func (o *ScCallTransfers) GetTypeId(keyId int32) int32 {
	return OBJTYPE_INT
}

func (o *ScCallTransfers) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Transfers = make(map[balance.Color]int64)
	default:
		var color balance.Color = [32]byte{}
		copy(color[:], o.vm.getKeyFromId(keyId))
		o.Transfers[color] = value
	}
}
