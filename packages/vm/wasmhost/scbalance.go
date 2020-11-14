package wasmhost

import (
	"bytes"
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
	if o.requestOnly {
		request := o.vm.ctx.AccessRequest()
		reqId := o.vm.ctx.RequestID()
		if bytes.Equal(key, reqId.TransactionID().Bytes()) {
			return request.NumFreeMintedTokens()
		}
	}
	color, _, err := balance.ColorFromBytes(key)
	if err != nil {
		o.Error(err.Error())
		return 0
	}
	account := o.vm.ctx.AccessSCAccount()
	if o.requestOnly {
		return account.AvailableBalanceFromRequest(&color)
	}
	return account.AvailableBalance(&color)
}

func (o *ScBalance) GetTypeId(keyId int32) int32 {
	if o.Exists(keyId) {
		return OBJTYPE_INT
	}
	return -1
}
