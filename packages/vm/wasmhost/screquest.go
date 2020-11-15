package wasmhost

type ScRequest struct {
	MapObject
}

func (o *ScRequest) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
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
	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyColors:  func() WaspObject { return &ScColors{requestOnly: true} },
		KeyBalance: func() WaspObject { return &ScBalance{requestOnly: true} },
		KeyParams:  func() WaspObject { return &ScRequestParams{} },
	})
}

func (o *ScRequest) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyAddress:
		return OBJTYPE_BYTES
	case KeyBalance:
		return OBJTYPE_MAP
	case KeyColors:
		return OBJTYPE_BYTES_ARRAY
	case KeyHash:
		return OBJTYPE_BYTES
	case KeyId:
		return OBJTYPE_BYTES
	case KeyParams:
		return OBJTYPE_MAP
	case KeyTimestamp:
		return OBJTYPE_INT
	}
	return -1
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

//TODO keep track of field types
func (o *ScRequestParams) GetTypeId(keyId int32) int32 {
	return o.MapObject.GetTypeId(keyId)
}
