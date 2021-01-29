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
	description string
	name        string
	params      int32
}

func NewScDeployInfo(vm *wasmProcessor) *ScDeployInfo {
	o := &ScDeployInfo{}
	o.vm = vm
	return o
}

func (o *ScDeployInfo) Exists(keyId int32, typeId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScDeployInfo) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyDescription:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyHash:
		return wasmhost.OBJTYPE_HASH
	case wasmhost.KeyName:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyParams:
		return wasmhost.OBJTYPE_INT
	}
	return 0
}

func (o *ScDeployInfo) Invoke(programHash hashing.HashValue) {
	params := dict.New()
	if o.params != 0 {
		params = o.host.FindObject(o.params).(*ScDict).kvStore.(dict.Dict)
		params.MustIterate("", func(key kv.Key, value []byte) bool {
			o.Trace("  PARAM '%s'", key)
			return true
		})
	}

	o.Trace("DEPLOY c'%s' f'%s'", o.name, o.description)
	err := o.vm.ctx.DeployContract(programHash, o.name, o.description, params)
	if err != nil {
		o.Panic("failed to deploy: %v", err)
	}
}

func (o *ScDeployInfo) SetBytes(keyId int32, typeId int32, bytes []byte) {
	switch keyId {
	case wasmhost.KeyDescription:
		o.description = string(bytes)
	case wasmhost.KeyHash:
		hash,err := hashing.HashValueFromBytes(bytes)
		if err != nil { o.Panic(err.Error())}
		o.Invoke(hash)
	case wasmhost.KeyName:
		o.name = string(bytes)
	case wasmhost.KeyParams:
		o.params = int32(o.MustInt64(bytes))
	default:
		o.invalidKey(keyId)
	}
}
