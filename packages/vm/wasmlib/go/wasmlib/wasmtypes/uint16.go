// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScUint16Length = 2

func DecodeUint16(dec *wasmcodec.WasmDecoder) uint16 {
	return uint16(dec.VluDecode(16))
}

func EncodeUint16(enc *wasmcodec.WasmEncoder, value uint16) {
	enc.VluEncode(uint64(value))
}

func Uint16FromBytes(buf []byte) uint16 {
	if buf == nil {
		return 0
	}
	if len(buf) != ScUint16Length {
		Panic("invalid Uint16 length")
	}
	return uint16(buf[0]) | uint16(buf[1])<<8
}

func BytesFromUint16(value uint16) []byte {
	return []byte{byte(value), byte(value >> 8)}
}

func StringFromUint16(value uint16) string {
	return strconv.FormatUint(uint64(value), 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint16 struct {
	proxy Proxy
}

func NewScImmutableUint16(proxy Proxy) ScImmutableUint16 {
	return ScImmutableUint16{proxy: proxy}
}

func (o ScImmutableUint16) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableUint16) String() string {
	return StringFromUint16(o.Value())
}

func (o ScImmutableUint16) Value() uint16 {
	return Uint16FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableUint16 struct {
	ScImmutableUint16
}

func NewScMutableUint16(proxy Proxy) ScMutableUint16 {
	return ScMutableUint16{ScImmutableUint16{proxy: proxy}}
}

func (o ScMutableUint16) Delete() {
	o.proxy.Delete()
}

func (o ScMutableUint16) SetValue(value uint16) {
	o.proxy.Set(BytesFromUint16(value))
}
