package wasmhost

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

type ScBalance struct {
	MapObject
	requestOnly bool
}

func (o *ScBalance) Exists(keyId int32) bool {
	return o.GetInt(keyId) != 0
}

func (o *ScBalance) GetInt(keyId int32) int64 {
	key := o.vm.WasmHost.GetKey(keyId)
	color, _, err := balance.ColorFromBytes(key)
	if err != nil {
		o.Error(err.Error())
		return 0
	}

	accounts := o.vm.ctx.Accounts()
	if o.requestOnly {
		return accounts.Incoming().Balance(color)
	}
	return accounts.Balance(color)
}

func (o *ScBalance) GetTypeId(keyId int32) int32 {
	if o.Exists(keyId) {
		return OBJTYPE_INT
	}
	return -1
}
