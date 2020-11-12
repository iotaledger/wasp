package wasmhost

type ScAccount struct {
	MapObject
}

func (o *ScAccount) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScAccount) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyBalance: func() WaspObject { return &ScBalance{} },
		KeyColors:  func() WaspObject { return &ScColors{} },
	})
}

func (o *ScAccount) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyBalance:
		return OBJTYPE_MAP
	case KeyColors:
		return OBJTYPE_INT_ARRAY
	}
	return -1
}
