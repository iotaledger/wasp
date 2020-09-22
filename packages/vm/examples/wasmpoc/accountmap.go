package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces/objtype"
)

type AccountMap struct {
	MapObject
	balanceId int32
	colorsId  int32
}

func NewAccountMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &AccountMap{MapObject: MapObject{vm: h, name: "Account"}}
}

func (o *AccountMap) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyColors:
		return o.checkedObjectId(&o.colorsId, NewColorsArray, typeId, objtype.OBJTYPE_INT_ARRAY)
	case KeyBalance:
		return o.checkedObjectId(&o.balanceId, NewBalanceMap, typeId, objtype.OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}
