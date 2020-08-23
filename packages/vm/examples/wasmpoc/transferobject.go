package wasmpoc

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wart/host/interfaces"
)

type TransferObject struct {
	vm        *wasmVMPocProcessor
	address   string
	amount    int64
	color     string
}

func NewTransferObject(h *wasmVMPocProcessor) *TransferObject {
	return &TransferObject{vm: h}
}

func (o *TransferObject) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyXferAmount:
		return o.amount
	default:
		o.vm.SetError("Invalid key")
	}
	return 0
}

func (o *TransferObject) GetObjectId(keyId int32, typeId int32) int32 {
	panic("implement Transfer.GetObjectId")
}

func (o *TransferObject) GetString(keyId int32) string {
	switch keyId {
	case KeyXferAddress:
		return o.address
	case KeyXferColor:
		return o.color
	default:
		o.vm.SetError("Invalid key")
	}
	return ""
}

func (o *TransferObject) Send(ctx interfaces.HostInterface) {
	o.vm.Logf("XFER SEND a%d a'%16s' c'%16s'", o.amount, o.address, o.color)
	addr, err := address.FromBase58(o.address)
	if err != nil {
		o.vm.ctx.Panic("MoveTokens failed 1")
	}

	// when no color specified default is ColorIOTA
	bytes := balance.ColorIOTA[:]
	if o.color != "" {
		bytes, err = hex.DecodeString(o.color)
		if err != nil {
			o.vm.ctx.Panic("MoveTokens failed 2")
		}
	}
	color, _, err := balance.ColorFromBytes(bytes)
	if err != nil {
		o.vm.ctx.Panic("MoveTokens failed 3")
	}

	if !o.vm.ctx.AccessSCAccount().MoveTokens(&addr, &color, o.amount) {
		o.vm.Logf("$$$$$$$$$$ something went wrong")
		o.vm.ctx.Panic("MoveTokens failed 4")
	}
}

func (o *TransferObject) SetInt(keyId int32, value int64) {
	o.vm.Logf("Transfer.SetInt k%d v%d", keyId, value)
	switch keyId {
	case interfaces.KeyLength:
		// clear transfer, tracker will still know about it
		// so maybe move it to an allocation pool for reuse
		o.address = ""
		o.color = ""
		o.amount = 0
	case KeyXferAmount:
		o.amount = value
	default:
		o.vm.SetError("Invalid key")
	}
}

func (o *TransferObject) SetString(keyId int32, value string) {
	o.vm.Logf("Transfer.SetString k%d v'%s'", keyId, value)
	switch keyId {
	case KeyXferAddress:
		o.address = value
	case KeyXferColor:
		o.color = value
	default:
		o.vm.SetError("Invalid key")
	}
}
