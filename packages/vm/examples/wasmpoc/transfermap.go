package wasmpoc

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
)

type TransferMap struct {
	MapObject
	address string
	amount  int64
	color   string
}

func NewTransferMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &TransferMap{MapObject: MapObject{vm: h, name: "Transfer"}}
}

func (o *TransferMap) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyAmount:
		return o.amount
	}
	return o.MapObject.GetInt(keyId)
}

func (o *TransferMap) GetString(keyId int32) string {
	switch keyId {
	case KeyAddress:
		return o.address
	case KeyColor:
		return o.color
	}
	return o.MapObject.GetString(keyId)
}

func (o *TransferMap) Send(ctx interfaces.HostInterface) {
	o.vm.Logf("XFER SEND a%d a'%16s' c'%16s'", o.amount, o.address, o.color)
	addr, err := address.FromBase58(o.address)
	if err != nil {
		o.vm.ctx.Panic("MoveTokens failed 1")
	}

	// when no color specified default is ColorIOTA
	bytes := balance.ColorIOTA[:]
	if o.color != "" && o.color != "iota" {
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

func (o *TransferMap) SetInt(keyId int32, value int64) {
	switch keyId {
	case interfaces.KeyLength:
		// clear transfer, tracker will still know about it
		// so maybe move it to an allocation pool for reuse
		o.address = ""
		o.color = ""
		o.amount = 0
	case KeyAmount:
		o.amount = value
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *TransferMap) SetString(keyId int32, value string) {
	switch keyId {
	case KeyAddress:
		o.address = value
	case KeyColor:
		o.color = value
	default:
		o.MapObject.SetString(keyId, value)
	}
}
