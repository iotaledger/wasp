package wasmhost

type ScAccount struct {
	MapObject
}

func (o *ScAccount) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyColors:  {OBJTYPE_INT_ARRAY, func() WaspObject { return &ScColors{} }},
		KeyBalance: {OBJTYPE_MAP, func() WaspObject { return &ScBalance{} }},
	})
}
