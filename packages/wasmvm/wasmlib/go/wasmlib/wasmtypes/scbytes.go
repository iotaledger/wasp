// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func BytesDecode(dec *WasmDecoder) []byte {
	return dec.Bytes()
}

func BytesEncode(enc *WasmEncoder, value []byte) {
	enc.Bytes(value)
}

func BytesFromBytes(buf []byte) []byte {
	return buf
}

func BytesToBytes(buf []byte) []byte {
	return buf
}

func BytesToString(value []byte) string {
	return base58Encode(value)
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
	return BytesToString(o.Value())
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
	o.proxy.Set(BytesToBytes(value))
}
