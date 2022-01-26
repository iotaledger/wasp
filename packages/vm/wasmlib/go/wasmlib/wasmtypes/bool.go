// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeBool(dec *WasmDecoder) bool {
	return dec.Byte() != 0
}

func EncodeBool(enc *WasmEncoder, value bool) {
	if value {
		enc.Byte(1)
		return
	}
	enc.Byte(0)
}

func BoolFromBytes(buf []byte) bool {
	if buf == nil {
		return false
	}
	if len(buf) != 1 {
		panic("invalid Bool length")
	}
	if buf[0] == 0x00 {
		return false
	}
	if buf[0] != 0xff {
		panic("invalid Bool value")
	}
	return true
}

func BytesFromBool(value bool) []byte {
	if value {
		return []byte{0xff}
	}
	return []byte{0x00}
}

func BoolToString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableBool struct {
	proxy Proxy
}

func NewScImmutableBool(proxy Proxy) ScImmutableBool {
	return ScImmutableBool{proxy: proxy}
}

func (o ScImmutableBool) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableBool) String() string {
	return BoolToString(o.Value())
}

func (o ScImmutableBool) Value() bool {
	return BoolFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableBool struct {
	ScImmutableBool
}

func NewScMutableBool(proxy Proxy) ScMutableBool {
	return ScMutableBool{ScImmutableBool{proxy: proxy}}
}

func (o ScMutableBool) Delete() {
	o.proxy.Delete()
}

func (o ScMutableBool) SetValue(value bool) {
	o.proxy.Set(BytesFromBool(value))
}
