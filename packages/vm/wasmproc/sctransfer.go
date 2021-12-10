// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransfers struct {
	ScSandboxObject
	wc *WasmContext
}

func NewScTransfers(wc *WasmContext) *ScTransfers {
	return &ScTransfers{wc: wc}
}

func (a *ScTransfers) GetObjectID(keyID, typeID int32) int32 {
	return GetArrayObjectID(a, keyID, typeID, func() WaspObject {
		return NewScTransferInfo(a.wc)
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransferInfo struct {
	ScSandboxObject
	address ledgerstate.Address
	wc      *WasmContext
}

func NewScTransferInfo(wc *WasmContext) *ScTransferInfo {
	return &ScTransferInfo{wc: wc}
}

// TODO refactor
func (o *ScTransferInfo) Invoke(balances int32) {
	transfer := colored.NewBalances()
	balancesObj := o.host.FindObject(balances).(*ScDict)
	balancesObj.kvStore.MustIterate("", func(key kv.Key, value []byte) bool {
		if len(key) != ledgerstate.ColorLength {
			return true
		}
		col, err := codec.DecodeColor([]byte(key))
		if err != nil {
			o.Panicf(err.Error())
		}
		amount, err := codec.DecodeUint64(value)
		if err != nil {
			o.Panicf(err.Error())
		}
		o.Tracef("TRANSFER #%d c'%s' a'%s'", value, col.String(), o.address.Base58())
		transfer.Set(col, amount)
		return true
	})
	if !o.wc.ctx.Send(o.address, transfer, nil) {
		o.Panicf("failed to send to %s", o.address.Base58())
	}
}

func (o *ScTransferInfo) SetBytes(keyID, typeID int32, bytes []byte) {
	switch keyID {
	case wasmhost.KeyAddress:
		var err error
		o.address, _, err = ledgerstate.AddressFromBytes(bytes)
		if err != nil {
			o.Panicf("SetBytes: invalid address: " + err.Error())
		}
	case wasmhost.KeyBalances:
		balanceMapID, err := codec.DecodeInt32(bytes, 0)
		if err != nil {
			o.Panicf("SetBytes: invalid balance map id: " + err.Error())
		}
		o.Invoke(balanceMapID)
	default:
		o.InvalidKey(keyID)
	}
}
