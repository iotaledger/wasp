package wasmhost

import "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"

type ScColors struct {
	ArrayObject
	requestOnly bool
	colors      []balance.Color
}

func (a *ScColors) Exists(keyId int32) bool {
	return uint32(keyId) < uint32(a.GetLength())
}

func (a *ScColors) GetBytes(keyId int32) []byte {
	if a.Exists(keyId) {
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

func (a *ScColors) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return OBJTYPE_BYTES
	}
	return -1
}

func (a *ScColors) loadColors() {
	if a.colors != nil {
		return
	}

	//TODO for now we assume the colors are at least ColorIOTA and the colors in the request
	accounts := a.vm.ctx.Accounts()
	a.colors = append(a.colors, balance.ColorIOTA)
	accounts.Incoming().IterateDeterministic(func(color balance.Color, amount int64) bool {
		if color != balance.ColorIOTA {
			a.colors = append(a.colors, color)
		}
		return true
	})
}
