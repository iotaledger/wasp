// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransfers struct {
	ScSandboxObject
}

func NewScTransfers(vm *wasmProcessor) *ScTransfers {
	a := &ScTransfers{}
	a.vm = vm
	return a
}

func (a *ScTransfers) GetObjectID(keyID int32, typeID int32) int32 {
	return GetArrayObjectID(a, keyID, typeID, func() WaspObject {
		return NewScTransferInfo(a.vm)
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransferInfo struct {
	ScSandboxObject
	address ledgerstate.Address
}

func NewScTransferInfo(vm *wasmProcessor) *ScTransferInfo {
	o := &ScTransferInfo{}
	o.vm = vm
	return o
}

// TODO refactor
func (o *ScTransferInfo) Invoke(balances int32) {
	balancesMap := make(map[ledgerstate.Color]uint64)
	balancesObj := o.host.FindObject(balances).(*ScDict)
	balancesObj.kvStore.MustIterate("", func(key kv.Key, value []byte) bool {
		if len(key) != ledgerstate.ColorLength {
			return true
		}
		color, _, err := codec.DecodeColor([]byte(key))
		if err != nil {
			o.Panic(err.Error())
		}
		amount, _, err := codec.DecodeUint64(value)
		if err != nil {
			o.Panic(err.Error())
		}
		o.Trace("TRANSFER #%d c'%s' a'%s'", value, color.String(), o.address.Base58())
		balancesMap[color] = amount
		return true
	})
	transfer := ledgerstate.NewColoredBalances(balancesMap)
	if !o.vm.ctx.Send(o.address, transfer, nil) {
		o.Panic("failed to send to %s", o.address.Base58())
	}
}

func (o *ScTransferInfo) SetBytes(keyID int32, typeID int32, bytes []byte) {
	switch keyID {
	case wasmhost.KeyAddress:
		var err error
		o.address, _, err = ledgerstate.AddressFromBytes(bytes)
		if err != nil {
			o.Panic("SetBytes: invalid address: " + err.Error())
		}
	case wasmhost.KeyBalances:
		balanceMapID, _, err := codec.DecodeInt32(bytes)
		if err != nil {
			o.Panic("SetBytes: invalid balance map id: " + err.Error())
		}
		o.Invoke(balanceMapID)
	default:
		o.invalidKey(keyID)
	}
}
