// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func StringDecode(dec *WasmDecoder) string {
	return string(dec.Bytes())
}

func StringEncode(enc *WasmEncoder, value string) {
	enc.Bytes([]byte(value))
}

func StringFromBytes(buf []byte) string {
	return string(buf)
}

func StringToBytes(value string) []byte {
	return []byte(value)
}

func StringToString(value string) string {
	return value
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableString struct {
	proxy Proxy
}

func NewScImmutableString(proxy Proxy) ScImmutableString {
	return ScImmutableString{proxy: proxy}
}

func (o ScImmutableString) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableString) String() string {
	return StringToString(o.Value())
}

func (o ScImmutableString) Value() string {
	return StringFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableString struct {
	ScImmutableString
}

func NewScMutableString(proxy Proxy) ScMutableString {
	return ScMutableString{ScImmutableString{proxy: proxy}}
}

func (o ScMutableString) Delete() {
	o.proxy.Delete()
}

func (o ScMutableString) SetValue(value string) {
	o.proxy.Set(StringToBytes(value))
}
