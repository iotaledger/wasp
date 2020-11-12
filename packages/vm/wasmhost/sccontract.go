package wasmhost

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

type ScContract struct {
	MapObject
}

func (o *ScContract) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScContract) GetBytes(keyId int32) []byte {
	switch keyId {
	case KeyAddress:
		id := o.vm.ctx.GetContractID()
		return id[:coretypes.ChainIDLength]
	case KeyColor: //TODO
	case KeyId:
		id := o.vm.ctx.GetContractID()
		return id[:]
	case KeyOwner:
		address := o.vm.ctx.GetOwnerAddress()
		return address[:]
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
	case KeyAddress:
		return OBJTYPE_BYTES
	case KeyColor:
		return OBJTYPE_BYTES
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
