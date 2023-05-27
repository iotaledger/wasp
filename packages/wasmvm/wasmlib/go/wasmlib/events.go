// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type ContractEvent struct {
	ChainID    wasmtypes.ScChainID
	ContractID wasmtypes.ScHname
	Topic      string
	Timestamp  uint64
	Payload    []byte
}

type IEventHandlers interface {
	CallHandler(topic string, dec *wasmtypes.WasmDecoder)
	ID() uint32
}

var nextID = uint32(0)

func EventHandlersGenerateID() uint32 {
	nextID++
	return nextID
}

func NewEventEncoder(topic string) *wasmtypes.WasmEncoder {
	enc := wasmtypes.NewWasmEncoder()
	wasmtypes.StringEncode(enc, topic)
	return enc
}

func EventEmit(enc *wasmtypes.WasmEncoder) {
	ScFuncContext{}.Event(enc.Buf())
}
