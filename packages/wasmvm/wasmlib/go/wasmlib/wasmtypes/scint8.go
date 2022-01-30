// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"strconv"
)

const ScInt8Length = 1

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func Int8Decode(dec *WasmDecoder) int8 {
	return int8(dec.Byte())
}

func Int8Encode(enc *WasmEncoder, value int8) {
	enc.Byte(byte(value))
}

func Int8FromBytes(buf []byte) int8 {
	if buf == nil {
		return 0
	}
	if len(buf) != ScInt8Length {
		panic("invalid Int8 length")
	}
	return int8(buf[0])
}

func Int8ToBytes(value int8) []byte {
	return []byte{byte(value)}
}

func Int8ToString(value int8) string {
	return strconv.FormatInt(int64(value), 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt8 struct {
	proxy Proxy
}

func NewScImmutableInt8(proxy Proxy) ScImmutableInt8 {
	return ScImmutableInt8{proxy: proxy}
}

func (o ScImmutableInt8) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableInt8) String() string {
	return Int8ToString(o.Value())
}

func (o ScImmutableInt8) Value() int8 {
	return Int8FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableInt8 struct {
	ScImmutableInt8
}

func NewScMutableInt8(proxy Proxy) ScMutableInt8 {
	return ScMutableInt8{ScImmutableInt8{proxy: proxy}}
}

func (o ScMutableInt8) Delete() {
	o.proxy.Delete()
}

func (o ScMutableInt8) SetValue(value int8) {
	o.proxy.Set(Int8ToBytes(value))
}
