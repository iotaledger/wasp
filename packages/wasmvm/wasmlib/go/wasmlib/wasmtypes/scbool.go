// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

const (
	ScBoolLength = 1
	ScBoolFalse  = 0x00
	ScBoolTrue   = 0xff
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func BoolDecode(dec *WasmDecoder) bool {
	return dec.Byte() != ScBoolFalse
}

func BoolEncode(enc *WasmEncoder, value bool) {
	if value {
		enc.Byte(ScBoolTrue)
		return
	}
	enc.Byte(ScBoolFalse)
}

func BoolFromBytes(buf []byte) bool {
	if buf == nil {
		return false
	}
	if len(buf) != ScBoolLength {
		panic("invalid Bool length")
	}
	if buf[0] == ScBoolFalse {
		return false
	}
	if buf[0] != ScBoolTrue {
		panic("invalid Bool value")
	}
	return true
}

func BoolToBytes(value bool) []byte {
	if value {
		return []byte{ScBoolTrue}
	}
	return []byte{ScBoolFalse}
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
	o.proxy.Set(BoolToBytes(value))
}
