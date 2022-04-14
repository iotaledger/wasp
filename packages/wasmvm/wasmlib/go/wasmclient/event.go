// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type Event struct {
	message   []string
	Timestamp uint32
}

func (e *Event) Init(message []string) {
	e.message = message
	e.Timestamp = e.NextUint32()
}

func (e *Event) next() string {
	next := e.message[0]
	e.message = e.message[1:]
	return next
}

func (e *Event) NextAddress() wasmtypes.ScAddress {
	return wasmtypes.AddressFromBytes(e.NextBytes())
}

func (e *Event) NextAgentID() wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(e.NextBytes())
}

func (e *Event) NextBytes() []byte {
	return Base58Decode(e.next())
}

func (e *Event) NextBool() bool {
	return e.next() != "0"
}

func (e *Event) NextChainID() wasmtypes.ScChainID {
	return wasmtypes.ChainIDFromBytes(e.NextBytes())
}

func (e *Event) NextHash() wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(e.NextBytes())
}

func (e *Event) NextHname() wasmtypes.ScHname {
	return wasmtypes.ScHname(e.nextUint(32))
}

func (e *Event) nextInt(bitSize int) int64 {
	val, err := strconv.ParseInt(e.next(), 10, bitSize)
	if err != nil {
		panic("int parse error")
	}
	return val
}

func (e *Event) NextInt8() int8 {
	return int8(e.nextInt(8))
}

func (e *Event) NextInt16() int16 {
	return int16(e.nextInt(16))
}

func (e *Event) NextInt32() int32 {
	return int32(e.nextInt(32))
}

func (e *Event) NextInt64() int64 {
	return e.nextInt(64)
}

func (e *Event) NextRequestID() wasmtypes.ScRequestID {
	return wasmtypes.RequestIDFromBytes(e.NextBytes())
}

func (e *Event) NextString() string {
	return e.next()
}

func (e *Event) nextUint(bitSize int) uint64 {
	val, err := strconv.ParseUint(e.next(), 10, bitSize)
	if err != nil {
		panic("uint parse error")
	}
	return val
}

func (e *Event) NextUint8() uint8 {
	return uint8(e.nextUint(8))
}

func (e *Event) NextUint16() uint16 {
	return uint16(e.nextUint(16))
}

func (e *Event) NextUint32() uint32 {
	return uint32(e.nextUint(32))
}

func (e *Event) NextUint64() uint64 {
	return e.nextUint(64)
}
