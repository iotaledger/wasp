// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
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

func (a *ScTransfers) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return NewScTransferInfo(a.vm)
	})
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransferInfo struct {
	ScSandboxObject
	address address.Address
}

func NewScTransferInfo(vm *wasmProcessor) *ScTransferInfo {
	o := &ScTransferInfo{}
	o.vm = vm
	return o
}

func (o *ScTransferInfo) Invoke(balances int32) {
	var transfer coretypes.ColoredBalances
	balancesObj := o.host.FindObject(balances)
	switch balancesObj.(type) {
	case *ScDict:
		balancesMap := make(map[balance.Color]int64)
		balancesDict := balancesObj.(*ScDict).kvStore
		balancesDict.MustIterate("", func(key kv.Key, value []byte) bool {
			color, _, err := codec.DecodeColor([]byte(key))
			if err != nil {
				o.Panic(err.Error())
			}
			amount, _, err := codec.DecodeInt64(value)
			if err != nil {
				o.Panic(err.Error())
			}
			o.Trace("TRANSFER #%d c'%s' a'%s'", value, color.String(), o.address.String())
			balancesMap[color] = amount
			return true
		})
		transfer = cbalances.NewFromMap(balancesMap)
	case *ScBalances:
		scBalances := balancesObj.(*ScBalances)
		if scBalances.incoming {
			transfer = o.vm.ctx.IncomingTransfer()
			break
		}
		transfer = o.vm.balances()
	default:
		o.Panic("Unexpected object type")
	}
	if !o.vm.ctx.TransferToAddress(o.address, transfer) {
		o.Panic("failed to transfer to %s", o.address.String())
	}
}

func (o *ScTransferInfo) SetBytes(keyId int32, value []byte) {
	var err error
	switch keyId {
	case wasmhost.KeyAddress:
		o.address, _, err = address.FromBytes(value)
		if err != nil {
			o.Panic("SetBytes: invalid address: " + err.Error())
		}
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScTransferInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyBalances:
		o.Invoke(int32(value))
	default:
		o.invalidKey(keyId)
	}
}
