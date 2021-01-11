// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCalls struct {
	ScSandboxObject
}

func NewScCalls(vm *wasmProcessor) *ScCalls {
	o := &ScCalls{}
	o.vm = vm
	return o
}

func (a *ScCalls) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return NewScCallInfo(a.vm)
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCallInfo struct {
	ScSandboxObject
	contract string
	function string
}

func NewScCallInfo(vm *wasmProcessor) *ScCallInfo {
	o := &ScCallInfo{}
	o.vm = vm
	return o
}

func (o *ScCallInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScCallInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		wasmhost.KeyParams:    func() WaspObject { return NewScDict(o.vm) },
		wasmhost.KeyResults:   func() WaspObject { return NewScDict(o.vm) },
		wasmhost.KeyTransfers: func() WaspObject { return NewScCallTransfers(o.vm) },
	})
}

func (o *ScCallInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyContract:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyDelay:
		return wasmhost.OBJTYPE_INT
	case wasmhost.KeyFunction:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyParams:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyResults:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyTransfers:
		return wasmhost.OBJTYPE_MAP
	}
	return 0
}

func (o *ScCallInfo) Invoke() {
	o.Trace("CALL c'%s' f'%s'", o.contract, o.function)
	contractCode := o.vm.contractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	paramsId := o.GetObjectId(wasmhost.KeyParams, wasmhost.OBJTYPE_MAP)
	params := o.host.FindObject(paramsId).(*ScDict).kvStore.(dict.Dict)
	params.MustIterate("", func(key kv.Key, value []byte) bool {
		o.Trace("  PARAM '%s'", key)
		return true
	})
	transfersId := o.GetObjectId(wasmhost.KeyTransfers, wasmhost.OBJTYPE_MAP)
	transfers := o.host.FindObject(transfersId).(*ScCallTransfers).Transfers
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
	resultsId := o.GetObjectId(wasmhost.KeyResults, wasmhost.OBJTYPE_MAP)
	o.host.FindObject(resultsId).(*ScDict).kvStore = results
}

func (o *ScCallInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		o.contract = ""
		o.function = ""
	case wasmhost.KeyDelay:
		if value != -1 {
			o.Panic("Unexpected delay: %d", value)
		}
		o.Invoke()
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScCallInfo) SetString(keyId int32, value string) {
	switch keyId {
	case wasmhost.KeyContract:
		o.contract = value
	case wasmhost.KeyFunction:
		o.function = value
	default:
		o.invalidKey(keyId)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScCallTransfers struct {
	ScSandboxObject
	Transfers map[balance.Color]int64
}

func NewScCallTransfers(vm *wasmProcessor) *ScCallTransfers {
	o := &ScCallTransfers{}
	o.vm = vm
	return o
}

func (o *ScCallTransfers) InitObj(id int32, keyId int32, owner *ScDict) {
	o.ScSandboxObject.InitObj(id, keyId, owner)
	o.Transfers = make(map[balance.Color]int64)
}

func (o *ScCallTransfers) Exists(keyId int32) bool {
	var color balance.Color
	copy(color[:], o.host.GetKeyFromId(keyId))
	return o.Transfers[color] != 0
}

func (o *ScCallTransfers) GetTypeId(keyId int32) int32 {
	return wasmhost.OBJTYPE_INT
}

func (o *ScCallTransfers) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		o.Transfers = make(map[balance.Color]int64)
	default:
		var color balance.Color
		copy(color[:], o.host.GetKeyFromId(keyId))
		o.Transfers[color] = value
	}
}
