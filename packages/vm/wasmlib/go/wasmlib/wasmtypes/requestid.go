// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package wasmtypes

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScRequestIDLength = 34

type ScRequestID struct {
	id [ScRequestIDLength]byte
}

func DecodeRequestID(dec *wasmcodec.WasmDecoder) ScRequestID {
	return newRequestIDFromBytes(dec.FixedBytes(ScRequestIDLength))
}

func EncodeRequestID(enc *wasmcodec.WasmEncoder, value ScRequestID) {
	enc.FixedBytes(value.Bytes(), ScRequestIDLength)
}

func RequestIDFromBytes(buf []byte) ScRequestID {
	if len(buf) != ScRequestIDLength {
		Panic("invalid RequestID length")
	}
	return newRequestIDFromBytes(buf)
}

func newRequestIDFromBytes(buf []byte) ScRequestID {
	o := ScRequestID{}
	copy(o.id[:], buf)
	return o
}

func (o ScRequestID) Bytes() []byte {
	return o.id[:]
}

func (o ScRequestID) String() string {
	// TODO standardize human readable string
	return base58Encode(o.id[:])
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
	return o.Value().String()
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
	o.proxy.Set(value.Bytes())
}
