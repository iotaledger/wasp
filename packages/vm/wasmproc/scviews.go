// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScViews struct {
	ScSandboxObject
}

func NewScViews(vm *wasmProcessor) *ScViews {
	o := &ScViews{}
	o.vm = vm
	return o
}

func (a *ScViews) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return NewScViewInfo(a.vm)
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScViewInfo struct {
	ScSandboxObject
	contract string
	function string
}

func NewScViewInfo(vm *wasmProcessor) *ScViewInfo {
	o := &ScViewInfo{}
	o.vm = vm
	return o
}

func (o *ScViewInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScViewInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		wasmhost.KeyParams:  func() WaspObject { return NewScDict(o.vm) },
		wasmhost.KeyResults: func() WaspObject { return NewScDict(o.vm) },
	})
}

func (o *ScViewInfo) GetTypeId(keyId int32) int32 {
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
	}
	return 0
}

func (o *ScViewInfo) Invoke() {
	o.Trace("VIEW c'%s' f'%s'", o.contract, o.function)
	contractCode := o.vm.contractID().Hname()
	if o.contract != "" {
		contractCode = coretypes.Hn(o.contract)
	}
	functionCode := coretypes.Hn(o.function)
	params := dict.New()
	paramsId, ok := o.objects[wasmhost.KeyParams]
	if ok {
		params = o.host.FindObject(paramsId).(*ScDict).kvStore.(dict.Dict)
		params.MustIterate("", func(key kv.Key, value []byte) bool {
			o.Trace("  PARAM '%s'", key)
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
	resultsId := o.GetObjectId(wasmhost.KeyResults, wasmhost.OBJTYPE_MAP)
	o.host.FindObject(resultsId).(*ScDict).kvStore = results
}

func (o *ScViewInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		o.contract = ""
		o.function = ""
	case wasmhost.KeyDelay:
		if value != -2 {
			o.Panic("Unexpected value for delay: %d", value)
		}
		o.Invoke()
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScViewInfo) SetString(keyId int32, value string) {
	switch keyId {
	case wasmhost.KeyContract:
		o.contract = value
	case wasmhost.KeyFunction:
		o.function = value
	default:
		o.invalidKey(keyId)
	}
}
