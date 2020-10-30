package wasmhost

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

type ScContract struct {
	MapObject
}

func (o *ScContract) Exists(keyId int32) bool {
	switch keyId {
	case KeyAddress:
	case KeyColor:
	case KeyDescription:
	case KeyId:
	case KeyName:
	case KeyOwner:
	default:
		return false
	}
	return true
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
