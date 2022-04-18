// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package wasmtypes

import (
	"strconv"
)

const ScInt16Length = 2

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func Int16Decode(dec *WasmDecoder) int16 {
	return int16(dec.VliDecode(16))
}

func Int16Encode(enc *WasmEncoder, value int16) {
	enc.VliEncode(int64(value))
}

func Int16FromBytes(buf []byte) int16 {
	if len(buf) == 0 {
		return 0
	}
	if len(buf) != ScInt16Length {
		panic("invalid Int16 length")
	}
	return int16(buf[0]) | int16(buf[1])<<8
}

func Int16ToBytes(value int16) []byte {
	return []byte{byte(value), byte(value >> 8)}
}

func Int16FromString(value string) int16 {
	return int16(IntFromString(value, 16))
}

func Int16ToString(value int16) string {
	return strconv.FormatInt(int64(value), 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt16 struct {
	proxy Proxy
}

func NewScImmutableInt16(proxy Proxy) ScImmutableInt16 {
	return ScImmutableInt16{proxy: proxy}
}

func (o ScImmutableInt16) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableInt16) String() string {
	return Int16ToString(o.Value())
}

func (o ScImmutableInt16) Value() int16 {
	return Int16FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableInt16 struct {
	ScImmutableInt16
}

func NewScMutableInt16(proxy Proxy) ScMutableInt16 {
	return ScMutableInt16{ScImmutableInt16{proxy: proxy}}
}

func (o ScMutableInt16) Delete() {
	o.proxy.Delete()
}

func (o ScMutableInt16) SetValue(value int16) {
	o.proxy.Set(Int16ToBytes(value))
}
