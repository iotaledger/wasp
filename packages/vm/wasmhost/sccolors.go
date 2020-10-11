package wasmhost

type ScColors struct {
	ArrayObject
	requestOnly bool
	colors      []string
}

func (a *ScColors) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.GetLength())
	}
	return a.ArrayObject.GetInt(keyId)
}

func (a *ScColors) GetLength() int32 {
	a.loadColors()
	return int32(len(a.colors))
}

func (a *ScColors) GetString(keyId int32) string {
	if keyId >= 0 && keyId < a.GetLength() {
		return a.colors[keyId]
	}
	return a.ArrayObject.GetString(keyId)
}

func (a *ScColors) loadColors() {
	if a.colors != nil {
		return
	}
	//TODO determine valid colors for account or request and add them base58-encoded to colors array
}
