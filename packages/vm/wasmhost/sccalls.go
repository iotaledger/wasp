// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type ScCallInfo struct {
	ScDict
	contract string
	function string
}

func (o *ScCallInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScCallInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetScDictId(o, keyId, typeId, ObjFactories{
		KeyParams:    func() WaspObject { return &ScDict{} },
		KeyResults:   func() WaspObject { return &ScDict{} },
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
	contractCode := o.vm.ContractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	paramsId := o.GetObjectId(KeyParams, OBJTYPE_MAP)
	params := o.vm.FindObject(paramsId).(*ScDict).Dict.(dict.Dict)
	params.MustIterate("", func(key kv.Key, value []byte) bool {
		o.vm.Trace("  PARAM '%s'", key)
		return true
	})
	transfersId := o.GetObjectId(KeyTransfers, OBJTYPE_MAP)
	transfers := o.vm.FindObject(transfersId).(*ScCallTransfers).Transfers
	balances := cbalances.NewFromMap(transfers)
	var err error
	var results dict.Dict
	if o.vm.ctx != nil {
		results, err = o.vm.ctx.Call(contractCode, functionCode, params, balances)
	} else {
		results, err = o.vm.ctxView.Call(contractCode, functionCode, params)
	}
	if err != nil {
		o.Panic("failed to invoke call: %v", err)
	}
	resultsId := o.GetObjectId(KeyResults, OBJTYPE_MAP)
	o.vm.FindObject(resultsId).(*ScDict).Dict = results
}

func (o *ScCallInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = ""
		o.function = ""
	case KeyDelay:
		if value != -1 {
			o.Panic("Unexpected value for delay: %d", value)
		}
		o.Invoke()
	default:
		o.ScDict.SetInt(keyId, value)
	}
}

func (o *ScCallInfo) SetString(keyId int32, value string) {
	switch keyId {
	case KeyContract:
		o.contract = value
	case KeyFunction:
		o.function = value
	default:
		o.ScDict.SetString(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScPostInfo struct {
	ScDict
	chainId  *coretypes.ChainID
	contract string
	delay    uint32
	function string
}

func (o *ScPostInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScPostInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetScDictId(o, keyId, typeId, ObjFactories{
		KeyParams:    func() WaspObject { return &ScDict{} },
		KeyTransfers: func() WaspObject { return &ScCallTransfers{} },
	})
}

func (o *ScPostInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyChain:
		return OBJTYPE_BYTES //TODO OBJTYPE_ADDRESS
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
	contractCode := o.vm.ContractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	params := dict.New()
	paramsId, ok := o.objects[KeyParams]
	if ok {
		params = o.vm.FindObject(paramsId).(*ScDict).Dict.(dict.Dict)
		params.MustIterate("", func(key kv.Key, value []byte) bool {
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
		o.Panic("failed to invoke post")
	}
}

func (o *ScPostInfo) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyChain:
		chainId, err := coretypes.NewChainIDFromBytes(value)
		if err != nil {
			o.Panic(err.Error())
		}
		o.chainId = &chainId
	default:
		o.ScDict.SetBytes(keyId, value)
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
			o.Panic("Unexpected value for delay: %d", value)
		}
		o.delay = uint32(value)
		o.Invoke()
	default:
		o.ScDict.SetInt(keyId, value)
	}
}

func (o *ScPostInfo) SetString(keyId int32, value string) {
	switch keyId {
	case KeyContract:
		o.contract = value
	case KeyFunction:
		o.function = value
	default:
		o.ScDict.SetString(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScViewInfo struct {
	ScDict
	contract string
	function string
}

func (o *ScViewInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScViewInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetScDictId(o, keyId, typeId, ObjFactories{
		KeyParams:  func() WaspObject { return &ScDict{} },
		KeyResults: func() WaspObject { return &ScDict{} },
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
	contractCode := o.vm.ContractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	params := dict.New()
	paramsId, ok := o.objects[KeyParams]
	if ok {
		params = o.vm.FindObject(paramsId).(*ScDict).Dict.(dict.Dict)
		params.MustIterate("", func(key kv.Key, value []byte) bool {
			o.vm.Trace("  PARAM '%s'", key)
			return true
		})
	}
	var err error
	var results dict.Dict
	if o.vm.ctx != nil {
		results, err = o.vm.ctx.Call(contractCode, functionCode, params, nil)
	} else {
		results, err = o.vm.ctxView.Call(contractCode, functionCode, params)
	}
	if err != nil {
		o.Panic("failed to invoke view: %v", err)
	}
	resultsId := o.GetObjectId(KeyResults, OBJTYPE_MAP)
	o.vm.FindObject(resultsId).(*ScDict).Dict = results
}

func (o *ScViewInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = ""
		o.function = ""
	case KeyDelay:
		if value != -2 {
			o.Panic("Unexpected value for delay: %d", value)
		}
		o.Invoke()
	default:
		o.ScDict.SetInt(keyId, value)
	}
}

func (o *ScViewInfo) SetString(keyId int32, value string) {
	switch keyId {
	case KeyContract:
		o.contract = value
	case KeyFunction:
		o.function = value
	default:
		o.ScDict.SetString(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCalls struct {
	ScDict
}

func (a *ScCalls) InitObj(id int32, keyId int32, owner *ModelObject) {
	a.ScDict.InitObj(id, keyId, owner)
	a.typeId = OBJTYPE_ARRAY | OBJTYPE_MAP
}

func (a *ScCalls) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return &ScCallInfo{}
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScPosts struct {
	ScDict
}

func (a *ScPosts) InitObj(id int32, keyId int32, owner *ModelObject) {
	a.ScDict.InitObj(id, keyId, owner)
	a.typeId = OBJTYPE_ARRAY | OBJTYPE_MAP
}

func (a *ScPosts) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return &ScPostInfo{}
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScViews struct {
	ScDict
}

func (a *ScViews) InitObj(id int32, keyId int32, owner *ModelObject) {
	a.ScDict.InitObj(id, keyId, owner)
	a.typeId = OBJTYPE_ARRAY | OBJTYPE_MAP
}

func (a *ScViews) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return &ScViewInfo{}
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCallTransfers struct {
	ScDict
	Transfers map[balance.Color]int64
}

func (o *ScCallTransfers) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.ScDict.InitObj(id, keyId, owner)
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
