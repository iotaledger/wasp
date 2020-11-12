package wasmhost

type ScAccount struct {
	MapObject
}

func (o *ScAccount) Exists(keyId int32) bool {
	switch keyId {
	case KeyBalance:
	case KeyColors:
	default:
		return false
	}
	return true
}

func (o *ScAccount) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyBalance: {OBJTYPE_MAP, func() WaspObject { return &ScBalance{} }},
		KeyColors:  {OBJTYPE_INT_ARRAY, func() WaspObject { return &ScColors{} }},
	})
}
