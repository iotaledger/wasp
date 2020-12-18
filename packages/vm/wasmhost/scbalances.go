// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

type ScBalances struct {
	ScDict
	incoming bool
}

func (o *ScBalances) Exists(keyId int32) bool {
	return o.GetInt(keyId) != 0
}

func (o *ScBalances) GetBytes(keyId int32) []byte {
	key := o.vm.GetKeyFromId(keyId)
	color, _, err := balance.ColorFromBytes(key)
	if err != nil {
		o.Panic("GetBytes: %v", err)
	}
	if color != balance.ColorNew {
		o.Panic("GetBytes: Expected ColorNew")
	}
	id := o.vm.ctx.RequestID()
	return id[:32]
}

func (o *ScBalances) GetInt(keyId int32) int64 {
	key := o.vm.GetKeyFromId(keyId)
	color, _, err := balance.ColorFromBytes(key)
	if err != nil {
		o.Panic("GetInt: %v", err)
	}
	balances := o.vm.Balances()
	if o.incoming {
		if o.vm.ctx == nil {
			return 0
		}
		balances = o.vm.ctx.IncomingTransfer()
	}
	return balances.Balance(color)
}

func (o *ScBalances) GetTypeId(keyId int32) int32 {
	if o.Exists(keyId) {
		return OBJTYPE_INT
	}
	return 0
}
