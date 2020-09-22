package wasmpoc

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
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
	case "new":
		color = balance.ColorNew
	default:
		if len(key) != 64 {
			o.error("GetInt: Invalid color key")
			return 0
		}
		bytes, err := hex.DecodeString(key)
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
