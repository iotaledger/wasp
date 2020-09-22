package wasmpoc

import "github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"

type ContractMap struct {
	MapObject
}

func NewContractMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &ContractMap{MapObject: MapObject{vm: h, name: "Contract"}}
}

func (o *ContractMap) GetString(keyId int32) string {
	switch keyId {
	case KeyAddress:
		return o.vm.ctx.GetSCAddress().String()
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
