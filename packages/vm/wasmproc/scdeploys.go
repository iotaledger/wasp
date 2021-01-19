package wasmproc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScDeploys struct {
	ScSandboxObject
}

func NewScDeploys(vm *wasmProcessor) *ScDeploys {
	a := &ScDeploys{}
	a.vm = vm
	return a
}

func (a *ScDeploys) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return NewScDeployInfo(a.vm)
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScDeployInfo struct {
	ScSandboxObject
	name        string
	description string
}

func NewScDeployInfo(vm *wasmProcessor) *ScDeployInfo {
	o := &ScDeployInfo{}
	o.vm = vm
	return o
}

func (o *ScDeployInfo) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScDeployInfo) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		wasmhost.KeyParams: func() WaspObject { return NewScDict(o.vm) },
	})
}

func (o *ScDeployInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyDescription:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyHash:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_HASH
	case wasmhost.KeyName:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyParams:
		return wasmhost.OBJTYPE_MAP
	}
	return 0
}

func (o *ScDeployInfo) Invoke(programHash []byte) {
	o.Trace("DEPLOY c'%s' f'%s'", o.name, o.description)
	paramsId := o.GetObjectId(wasmhost.KeyParams, wasmhost.OBJTYPE_MAP)
	params := o.host.FindObject(paramsId).(*ScDict).kvStore.(dict.Dict)
	params.MustIterate("", func(key kv.Key, value []byte) bool {
		o.Trace("  PARAM '%s'", key)
		return true
	})

	progHash, err := hashing.HashValueFromBytes(programHash)
	if err != nil {
		o.Panic("invalid hash: %v", err)
	}
	err = o.vm.ctx.DeployContract(progHash, o.name, o.description, params)
	if err != nil {
		o.Panic("failed to deploy: %v", err)
	}
}

func (o *ScDeployInfo) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case wasmhost.KeyHash:
		o.Invoke(value)
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScDeployInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		o.description = ""
		o.name = ""
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScDeployInfo) SetString(keyId int32, value string) {
	switch keyId {
	case wasmhost.KeyDescription:
		o.description = value
	case wasmhost.KeyName:
		o.name = value
	default:
		o.invalidKey(keyId)
	}
}
