package wasmhost

type ScContract struct {
	MapObject
}

func (o *ScContract) GetString(keyId int32) string {
	switch keyId {
	case KeyAddress:
		return o.vm.ctx.GetContractID().String()
	case KeyColor: //TODO
	case KeyDescription:
		return o.vm.GetDescription()
	case KeyId: //TODO
	case KeyName: //TODO
	case KeyOwner:
		return o.vm.ctx.GetOwnerAddress().String()
	}
	return o.MapObject.GetString(keyId)
}
