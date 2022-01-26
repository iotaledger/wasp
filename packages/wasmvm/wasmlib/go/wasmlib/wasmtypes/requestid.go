// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

const ScRequestIDLength = 34

type ScRequestID struct {
	id [ScRequestIDLength]byte
}

func requestIDFromBytes(buf []byte) ScRequestID {
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

func DecodeRequestID(dec *WasmDecoder) ScRequestID {
	return requestIDFromBytes(dec.FixedBytes(ScRequestIDLength))
}

func EncodeRequestID(enc *WasmEncoder, value ScRequestID) {
	enc.FixedBytes(value.Bytes(), ScRequestIDLength)
}

func RequestIDFromBytes(buf []byte) ScRequestID {
	if buf == nil {
		return ScRequestID{}
	}
	if len(buf) != ScRequestIDLength {
		panic("invalid RequestID length")
	}
	// final uint16 output index must be > ledgerstate.MaxOutputCount
	if buf[ScHashLength] > 127 || buf[ScHashLength+1] != 0 {
		panic("invalid RequestID: output index > 127")
	}
	return requestIDFromBytes(buf)
}

func BytesFromRequestID(value ScRequestID) []byte {
	return value.Bytes()
}

func RequestIDToString(value ScRequestID) string {
	return value.String()
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
