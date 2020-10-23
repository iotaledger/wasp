package wasmhost

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/mr-tron/base58/base58"
)

type ScBalance struct {
	MapObject
	requestOnly bool
}

func (o *ScBalance) GetInt(keyId int32) int64 {
	color := balance.ColorIOTA
	key := o.vm.GetKey(keyId)
	switch key {
	case "iota":
		color = balance.ColorIOTA
	default:
		if o.requestOnly {
			request := o.vm.ctx.AccessRequest()
			reqId := request.ID()
			if key == reqId.TransactionID().String() {
				return request.NumFreeMintedTokens()
			}
		}
		bytes, err := base58.Decode(key)
		if err != nil {
			panic(err)
		}
		color, _, err = balance.ColorFromBytes(bytes)
		if err != nil {
			panic(err)
		}
	}
	account := o.vm.ctx.AccessSCAccount()
	if o.requestOnly {
		return account.AvailableBalanceFromRequest(&color)
	}
	return account.AvailableBalance(&color)
}
