// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"encoding/binary"
	"strconv"
)

const ScUint32Length = 4

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func Uint32Decode(dec *WasmDecoder) uint32 {
	return uint32(dec.VluDecode(32))
}

func Uint32Encode(enc *WasmEncoder, value uint32) {
	enc.VluEncode(uint64(value))
}

func Uint32FromBytes(buf []byte) uint32 {
	if len(buf) == 0 {
		return 0
	}
	if len(buf) != ScUint32Length {
		panic("invalid Uint32 length")
	}
	return binary.LittleEndian.Uint32(buf)
}

func Uint32ToBytes(value uint32) []byte {
	tmp := make([]byte, ScUint32Length)
	binary.LittleEndian.PutUint32(tmp, value)
	return tmp
}

func Uint32FromString(value string) uint32 {
	return uint32(UintFromString(value, 32))
}

func Uint32ToString(value uint32) string {
	return strconv.FormatUint(uint64(value), 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint32 struct {
	proxy Proxy
}

func NewScImmutableUint32(proxy Proxy) ScImmutableUint32 {
	return ScImmutableUint32{proxy: proxy}
}

func (o ScImmutableUint32) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableUint32) String() string {
	return Uint32ToString(o.Value())
}

func (o ScImmutableUint32) Value() uint32 {
	return Uint32FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableUint32 struct {
	ScImmutableUint32
}

func NewScMutableUint32(proxy Proxy) ScMutableUint32 {
	return ScMutableUint32{ScImmutableUint32{proxy: proxy}}
}

func (o ScMutableUint32) Delete() {
	o.proxy.Delete()
}

func (o ScMutableUint32) SetValue(value uint32) {
	o.proxy.Set(Uint32ToBytes(value))
}
