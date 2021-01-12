// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScPosts struct {
	ScSandboxObject
}

func NewScPosts(vm *wasmProcessor) *ScPosts {
	o := &ScPosts{}
	o.vm = vm
	return o
}

func (a *ScPosts) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return NewScPostInfo(a.vm)
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScPostInfo struct {
	ScSandboxObject
	chainId  *coretypes.ChainID
	contract string
	delay    uint32
	function string
}

func NewScPostInfo(vm *wasmProcessor) *ScPostInfo {
	o := &ScPostInfo{}
	o.vm = vm
	return o
}

func (o *ScPostInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScPostInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		wasmhost.KeyParams:    func() WaspObject { return NewScDict(o.vm) },
		wasmhost.KeyTransfers: func() WaspObject { return NewScCallTransfers(o.vm) },
	})
}

func (o *ScPostInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyChain:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_ADDRESS
	case wasmhost.KeyContract:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyDelay:
		return wasmhost.OBJTYPE_INT
	case wasmhost.KeyFunction:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyParams:
		return wasmhost.OBJTYPE_MAP
	case wasmhost.KeyTransfers:
		return wasmhost.OBJTYPE_MAP
	}
	return 0
}

func (o *ScPostInfo) Invoke() {
	o.Trace("POST c'%s' f'%s' d%d", o.contract, o.function, o.delay)
	chainId := o.vm.ctx.ChainID()
	if o.chainId != nil {
		chainId = *o.chainId
	}
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
	transfersId := o.GetObjectId(wasmhost.KeyTransfers, wasmhost.OBJTYPE_MAP)
	transfers := o.host.FindObject(transfersId).(*ScCallTransfers).Transfers
	balances := cbalances.NewFromMap(transfers)
	if !o.vm.ctx.PostRequest(vmtypes.PostRequestParams{
		TargetContractID: coretypes.NewContractID(chainId, contractCode),
		EntryPoint:       functionCode,
		Params:           params,
		TimeLock:         util.NanoSecToUnixSec(o.vm.ctx.GetTimestamp()) + o.delay,
		Transfer:         balances,
	}) {
		o.Panic("failed to invoke post")
	}
}

func (o *ScPostInfo) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case wasmhost.KeyChain:
		chainId, err := coretypes.NewChainIDFromBytes(value)
		if err != nil {
			o.Panic(err.Error())
		}
		o.chainId = &chainId
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScPostInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		o.chainId = nil
		o.contract = ""
		o.delay = 0
		o.function = ""
	case wasmhost.KeyDelay:
		if value < 0 {
			o.Panic("Unexpected value for delay: %d", value)
		}
		o.delay = uint32(value)
		o.Invoke()
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScPostInfo) SetString(keyId int32, value string) {
	switch keyId {
	case wasmhost.KeyContract:
		o.contract = value
	case wasmhost.KeyFunction:
		o.function = value
	default:
		o.invalidKey(keyId)
	}
}
