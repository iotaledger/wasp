package wasmhost

type ScRequest struct {
	MapObject
}

func (o *ScRequest) Exists(keyId int32) bool {
	switch keyId {
	case KeyAddress:
	case KeyBalance:
	case KeyColors:
	case KeyHash:
	case KeyId:
	case KeyParams:
	case KeyTimestamp:
	default:
		return false
	}
	return true
}

func (o *ScRequest) GetBytes(keyId int32) []byte {
	switch keyId {
	case KeyAddress:
		return o.vm.ctx.AccessRequest().MustSenderAddress().Bytes()
	case KeyHash:
		id := o.vm.ctx.AccessRequest().ID()
		return id.TransactionID().Bytes()
	case KeyId:
		id := o.vm.ctx.AccessRequest().ID()
		return id.Bytes()
	}
	return o.MapObject.GetBytes(keyId)
}

func (o *ScRequest) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyTimestamp:
		return o.vm.ctx.GetTimestamp()
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScRequest) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyColors:  {OBJTYPE_INT_ARRAY, func() WaspObject { return &ScColors{requestOnly: true} }},
		KeyBalance: {OBJTYPE_MAP, func() WaspObject { return &ScBalance{requestOnly: true} }},
		KeyParams:  {OBJTYPE_MAP, func() WaspObject { return &ScRequestParams{} }},
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScRequestParams struct {
	MapObject
}

func (o *ScRequestParams) Exists(keyId int32) bool {
	key := o.vm.GetKey(keyId)
	exists, _ := o.vm.params.Has(key)
	return exists
}

func (o *ScRequestParams) GetBytes(keyId int32) []byte {
	key := o.vm.GetKey(keyId)
	value, _ := o.vm.params.Get(key)
	return value
}

func (o *ScRequestParams) GetInt(keyId int32) int64 {
	key := o.vm.GetKey(keyId)
	value, _, _ := o.vm.params.GetInt64(key)
	return value
}

func (o *ScRequestParams) GetString(keyId int32) string {
	key := o.vm.GetKey(keyId)
	value, _, _ := o.vm.params.GetString(key)
	return value
}
