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

func (o *ScTransfer) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) >= 0
}

func (o *ScTransfer) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyAddress:
		return OBJTYPE_BYTES
	case KeyColor:
		return OBJTYPE_BYTES
	case KeyAmount:
		return OBJTYPE_INT
	}
	return -1
}

func (o *ScTransfer) Send() {
	o.vm.Trace("TRANSFER a%d c'%s' a'%s'", o.amount, o.color.String(), o.address.String())

	// TODO
	o.vm.ctx.Panic("to be refactored")
	//if !o.vm.ctx.AccessSCAccount().MoveTokens(&o.address, &o.color, o.amount) {
	//	o.vm.ctx.Panic("Failed to move tokens")
	//}
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
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		transfer := &ScTransfer{}
		transfer.name = "transfer"
		return transfer
	})
}

func (a *ScTransfers) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return OBJTYPE_MAP
	}
	return -1
}

func (a *ScTransfers) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		a.objects = nil
	default:
		a.Error("SetInt: Invalid access")
	}
}

func (a *ScTransfers) SetString(keyId int32, value string) {
	a.Error("SetString: Invalid access")
}
