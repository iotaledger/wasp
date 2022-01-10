// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
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
	address iotago.Address
	wc      *WasmContext
}

func NewScTransferInfo(wc *WasmContext) *ScTransferInfo {
	return &ScTransferInfo{wc: wc}
}

// TODO refactor
func (o *ScTransferInfo) Invoke(balances int32) {
	transfer := iscp.NewEmptyAssets()
	balancesObj := o.host.FindObject(balances).(*ScDict)
	balancesObj.kvStore.MustIterate("", func(key kv.Key, value []byte) bool {
		panic("TODO implement - we need to support big int on wasm :/")
		// o.Tracef("TRANSFER #%d c'%s' a'%s'", value, new(big.Int).SetBytes(value).String(), o.address.Bech32(iscp.Bech32Prefix))
		// transfer.AddAsset([]byte(key), new(big.Int).SetBytes(value))
		// return true
	})
	if !o.wc.ctx.Send(o.address, transfer, nil) {
		o.Panicf("failed to send to %s", o.address.Base58())
	}
}

func (o *ScTransferInfo) SetBytes(keyID, typeID int32, bytes []byte) {
	panic("TODO implement")
	// switch keyID {
	// case wasmhost.KeyAddress:
	// 	var err error
	// 	o.address, _, err = iotago.AddressFromBytes(bytes)
	// 	if err != nil {
	// 		o.Panicf("SetBytes: invalid address: " + err.ErrorStr())
	// 	}
	// case wasmhost.KeyBalances:
	// 	balanceMapID, err := codec.DecodeInt32(bytes, 0)
	// 	if err != nil {
	// 		o.Panicf("SetBytes: invalid balance map id: " + err.ErrorStr())
	// 	}
	// 	o.Invoke(balanceMapID)
	// default:
	// 	o.InvalidKey(keyID)
	// }
}
