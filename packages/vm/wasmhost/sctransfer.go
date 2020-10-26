package wasmhost

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

type ScTransfer struct {
	MapObject
	address address.Address
	amount  int64
	color   balance.Color
}

func (o *ScTransfer) Send() {
	o.vm.Trace("TRANSFER a%d c'%16s' a'%16s'", o.amount, o.color.String(), o.address.String())
	if !o.vm.ctx.AccessSCAccount().MoveTokens(&o.address, &o.color, o.amount) {
		o.vm.ctx.Panic("Failed to move tokens")
	}
}

func (o *ScTransfer) SetBytes(keyId int32, value []byte) {
	var err error
	switch keyId {
	case KeyAddress:
		o.address, _, err = address.FromBytes(value)
		if err != nil {
			o.vm.ctx.Panic("Invalid address")
		}
	case KeyColor:
		o.color, _, err = balance.ColorFromBytes(value)
		if err != nil {
			o.vm.ctx.Panic("Invalid color")
		}
	default:
		o.MapObject.SetBytes(keyId, value)
	}
}

func (o *ScTransfer) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.address = address.Empty
		o.color = balance.ColorIOTA
		o.amount = 0
	case KeyAmount:
		o.amount = value
		o.Send()
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransfers struct {
	ArrayObject
}

func (a *ScTransfers) GetObjectId(keyId int32, typeId int32) int32 {
	return a.GetArrayObjectId(keyId, typeId, func() WaspObject {
		transfer := &ScTransfer{}
		transfer.name = "transfer"
		return transfer
	})
}

func (a *ScTransfers) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		a.objects = nil
	default:
		a.error("SetInt: Invalid access")
	}
}

func (a *ScTransfers) SetString(keyId int32, value string) {
	a.error("SetString: Invalid access")
}
