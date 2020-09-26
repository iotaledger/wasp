package wasmpoc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasplib/host/interfaces"
	"github.com/mr-tron/base58/base58"
)

type BalanceMap struct {
	MapObject
	requestOnly bool
}

func NewBalanceMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &BalanceMap{MapObject: MapObject{vm: h, name: "Balance"}, requestOnly: false}
}

func NewBalanceMapRequest(h *wasmVMPocProcessor) interfaces.HostObject {
	return &BalanceMap{MapObject: MapObject{vm: h, name: "Balance"}, requestOnly: true}
}

func (o *BalanceMap) GetInt(keyId int32) int64 {
	color := balance.ColorIOTA
	key := o.vm.GetKey(keyId)
	o.vm.Logf("Balance.GetInt: Key %d is '%s'", keyId, key)
	switch key {
	case "iota":
		color = balance.ColorIOTA
	case "new":
		color = balance.ColorNew
	default:
		if o.requestOnly {
			request := o.vm.ctx.AccessRequest()
			reqId := request.ID()
			if key == reqId.TransactionId().String() {
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
