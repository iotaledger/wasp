package wasmpoc

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

type BalanceObject struct {
	vm *wasmVMPocProcessor
	reqOnly bool
}

func NewBalanceObject(h *wasmVMPocProcessor, reqOnly bool) *BalanceObject {
	return &BalanceObject{vm: h, reqOnly: reqOnly }
}

func (o *BalanceObject) GetInt(keyId int32) int64 {
	if keyId <= 0 {
		o.vm.SetError("Invalid color key")
		return 0
	}
	color := balance.ColorIOTA
	key := o.vm.GetKey(keyId)
	switch key {
	case "iota":
	case "new":
		color = balance.ColorNew
	default:
		if len(key) != 64 {
			o.vm.SetError("Invalid color")
		}
		bytes,err := hex.DecodeString(key)
		if err != nil { panic(err) }
		color,_,err = balance.ColorFromBytes(bytes)
		if err != nil { panic(err) }
	}
	account := o.vm.ctx.AccessSCAccount()
	if o.reqOnly {
		return account.AvailableBalanceFromRequest(&color)
	}
	return account.AvailableBalance(&color)
}

func (o *BalanceObject) GetObjectId(keyId int32, typeId int32) int32 {
	panic("implement Balance.GetObjectId")
}

func (o *BalanceObject) GetString(keyId int32) string {
	panic("implement Balance.GetString")
}

func (o *BalanceObject) SetInt(keyId int32, value int64) {
	panic("implement Balance.SetInt")
}

func (o *BalanceObject) SetString(keyId int32, value string) {
	panic("implement Balance.SetString")
}
