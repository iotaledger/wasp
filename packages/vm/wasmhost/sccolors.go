package wasmhost

import "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"

type ScColors struct {
	ArrayObject
	requestOnly bool
	colors      []balance.Color
}

func (a *ScColors) Exists(keyId int32) bool {
	return keyId >= 0 && keyId < a.GetLength()
}

func (a *ScColors) GetBytes(keyId int32) []byte {
	if keyId >= 0 && keyId < a.GetLength() {
		return a.colors[keyId].Bytes()
	}
	return a.ArrayObject.GetBytes(keyId)
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

func (a *ScColors) loadColors() {
	if a.colors != nil {
		return
	}
	//TODO determine valid colors for account or request and add them base58-encoded to colors array
}
