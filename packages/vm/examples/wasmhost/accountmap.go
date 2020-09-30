package wasmhost

type AccountMap struct {
	MapObject
	balanceId int32
	colorsId  int32
}

func NewAccountMap(vm *wasmVMPocProcessor) HostObject {
	return &AccountMap{MapObject: MapObject{vm: vm, name: "Account"}}
}

func (o *AccountMap) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyColors:
		return o.checkedObjectId(&o.colorsId, NewColorsArray, typeId, OBJTYPE_INT_ARRAY)
	case KeyBalance:
		return o.checkedObjectId(&o.balanceId, NewBalanceMap, typeId, OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}
