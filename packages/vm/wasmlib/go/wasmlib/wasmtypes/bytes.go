// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeBytes(dec *wasmcodec.WasmDecoder) []byte {
	return dec.Bytes()
}

func EncodeBytes(enc *wasmcodec.WasmEncoder, value []byte) {
	enc.Bytes(value)
}

func BytesFromBytes(buf []byte) []byte {
	return buf
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableBytes struct {
	proxy Proxy
}

func NewScImmutableBytes(proxy Proxy) ScImmutableBytes {
	return ScImmutableBytes{proxy: proxy}
}

func (o ScImmutableBytes) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableBytes) String() string {
	return StringFromBytes(o.Value())
}

func (o ScImmutableBytes) Value() []byte {
	return BytesFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableBytes struct {
	ScImmutableBytes
}

func NewScMutableBytes(proxy Proxy) ScMutableBytes {
	return ScMutableBytes{ScImmutableBytes{proxy: proxy}}
}

func (o ScMutableBytes) Delete() {
	o.proxy.Delete()
}

func (o ScMutableBytes) SetValue(value []byte) {
	o.proxy.Set(BytesFromBytes(value))
}
