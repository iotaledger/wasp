// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

type ScContract struct {
	MapObject
}

func (o *ScContract) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScContract) GetBytes(keyId int32) []byte {
	switch keyId {
	case KeyId:
		id := o.vm.ContractID()
		return id[:]
	case KeyOwner:
		id := o.vm.ctx.ChainOwnerID()
		return id[:]
	}
	return o.MapObject.GetBytes(keyId)
}

func (o *ScContract) GetString(keyId int32) string {
	switch keyId {
	case KeyDescription:
		return o.vm.GetDescription()
	case KeyName: //TODO
	}
	return o.MapObject.GetString(keyId)
}

func (o *ScContract) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyDescription:
		return OBJTYPE_STRING
	case KeyId:
		return OBJTYPE_STRING
	case KeyName:
		return OBJTYPE_STRING
	case KeyOwner:
		return OBJTYPE_BYTES
	}
	return -1
}
