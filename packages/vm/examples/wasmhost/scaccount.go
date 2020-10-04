package wasmhost

type ScAccount struct {
	MapObject
	balanceId int32
	colorsId  int32
}

func (o *ScAccount) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyColors:
		return o.checkedObjectId(&o.colorsId, NewColorsArray, typeId, OBJTYPE_INT_ARRAY)
	case KeyBalance:
		return o.checkedObjectId(&o.balanceId, NewBalanceMap, typeId, OBJTYPE_MAP)
	default:
		return o.MapObject.GetObjectId(keyId, typeId)
	}
}
