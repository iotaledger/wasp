// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

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
	balances := a.vm.Balances()
	if a.requestOnly {
		if a.vm.ctx == nil {
			return
		}
		balances = a.vm.ctx.IncomingTransfer()
	}
	balances.IterateDeterministic(func(color balance.Color, amount int64) bool {
		a.colors = append(a.colors, color)
		return true
	})
}
