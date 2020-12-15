// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

type ScContract struct {
	ScDict
}

func (o *ScContract) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScContract) GetBytes(keyId int32) []byte {
	switch keyId {
	case KeyChain:
		id := o.vm.ctx.ChainID()
		return id[:]
	case KeyId:
		id := o.vm.ContractID()
		return id[:]
	case KeyOwner:
		id := o.vm.ctx.ChainOwnerID()
		return id[:]
	}
	return o.ScDict.GetBytes(keyId)
}

func (o *ScContract) GetString(keyId int32) string {
	switch keyId {
	case KeyDescription:
		return o.vm.GetDescription()
	case KeyName: //TODO
	}
	return o.ScDict.GetString(keyId)
}

func (o *ScContract) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyChain:
		return OBJTYPE_BYTES //TODO OBJTYPE_ADDRESS
	case KeyDescription:
		return OBJTYPE_STRING
	case KeyId:
		return OBJTYPE_BYTES
	case KeyName:
		return OBJTYPE_STRING
	case KeyOwner:
		return OBJTYPE_BYTES //TODO OBJTYPE_AGENT
	}
	return 0
}
