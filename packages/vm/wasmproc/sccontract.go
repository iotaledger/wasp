// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import "github.com/iotaledger/wasp/packages/vm/wasmhost"

type ScContract struct {
	ScDict
}

func (o *ScContract) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScContract) GetBytes(keyId int32) []byte {
	switch keyId {
	case wasmhost.KeyChain:
		id := o.vm.ctx.ChainID()
		return id[:]
	case wasmhost.KeyId:
		id := o.vm.ContractID()
		return id[:]
	case wasmhost.KeyOwner:
		id := o.vm.ctx.ChainOwnerID()
		return id[:]
	}
	return o.ScDict.GetBytes(keyId)
}

func (o *ScContract) GetString(keyId int32) string {
	switch keyId {
	case wasmhost.KeyDescription:
		return o.vm.GetDescription()
	case wasmhost.KeyName: //TODO
	}
	return o.ScDict.GetString(keyId)
}

func (o *ScContract) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyChain:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_ADDRESS
	case wasmhost.KeyDescription:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyId:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyName:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyOwner:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_AGENT
	}
	return 0
}
