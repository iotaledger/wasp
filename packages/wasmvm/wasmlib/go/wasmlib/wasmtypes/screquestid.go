// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"strings"

	"github.com/iotaledger/wasp/packages/util"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScRequestIDLength = 34
const RequestIDSeparator = "-"

type ScRequestID struct {
	id [ScRequestIDLength]byte
}

func (o ScRequestID) Bytes() []byte {
	return RequestIDToBytes(o)
}

func (o ScRequestID) String() string {
	return RequestIDToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func RequestIDDecode(dec *WasmDecoder) ScRequestID {
	return requestIDFromBytesUnchecked(dec.FixedBytes(ScRequestIDLength))
}

func RequestIDEncode(enc *WasmEncoder, value ScRequestID) {
	enc.FixedBytes(value.id[:], ScRequestIDLength)
}

func RequestIDFromBytes(buf []byte) ScRequestID {
	if len(buf) == 0 {
		return ScRequestID{}
	}
	if len(buf) != ScRequestIDLength {
		panic("invalid RequestID length")
	}
	// final uint16 output index must be > ledgerstate.MaxOutputCount
	if buf[ScRequestIDLength-2] > 127 || buf[ScRequestIDLength-1] != 0 {
		panic("invalid RequestID: output index > 127")
	}
	return requestIDFromBytesUnchecked(buf)
}

func RequestIDToBytes(value ScRequestID) []byte {
	return value.id[:]
}

func RequestIDFromString(value string) ScRequestID {
	elts := strings.Split(value, RequestIDSeparator)
	index := util.Uint16To2Bytes(Uint16FromString(elts[0]))
	buf := HexDecode(elts[1])
	return RequestIDFromBytes(append(buf, index...))
}

func RequestIDToString(value ScRequestID) string {
	reqID := RequestIDToBytes(value)
	// the last 2 byte is the TransactionOutputIndex
	txID := HexEncode(reqID[:ScRequestIDLength-2])
	index, _ := util.Uint16From2Bytes(reqID[ScRequestIDLength-2:])
	return Uint16ToString(index) + RequestIDSeparator + txID
}

func requestIDFromBytesUnchecked(buf []byte) ScRequestID {
	o := ScRequestID{}
	copy(o.id[:], buf)
	return o
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableRequestID struct {
	proxy Proxy
}

func NewScImmutableRequestID(proxy Proxy) ScImmutableRequestID {
	return ScImmutableRequestID{proxy: proxy}
}

func (o ScImmutableRequestID) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableRequestID) String() string {
	return RequestIDToString(o.Value())
}

func (o ScImmutableRequestID) Value() ScRequestID {
	return RequestIDFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableRequestID struct {
	ScImmutableRequestID
}

func NewScMutableRequestID(proxy Proxy) ScMutableRequestID {
	return ScMutableRequestID{ScImmutableRequestID{proxy: proxy}}
}

func (o ScMutableRequestID) Delete() {
	o.proxy.Delete()
}

func (o ScMutableRequestID) SetValue(value ScRequestID) {
	o.proxy.Set(RequestIDToBytes(value))
}
