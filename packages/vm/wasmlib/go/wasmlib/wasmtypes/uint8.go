// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeUint8(dec *wasmcodec.WasmDecoder) uint8 {
	return dec.Byte()
}

func EncodeUint8(enc *wasmcodec.WasmEncoder, value uint8) {
	enc.Byte(value)
}

func Uint8FromBytes(buf []byte) uint8 {
	if buf == nil {
		return 0
	}
	if len(buf) != 1 {
		Panic("invalid Uint8 length")
	}
	return buf[0]
}

func BytesFromUint8(value uint8) []byte {
	return []byte{value}
}

func StringFromUint8(value uint8) string {
	return strconv.FormatUint(uint64(value), 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint8 struct {
	proxy Proxy
}

func NewScImmutableUint8(proxy Proxy) ScImmutableUint8 {
	return ScImmutableUint8{proxy: proxy}
}

func (o ScImmutableUint8) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableUint8) String() string {
	return StringFromUint8(o.Value())
}

func (o ScImmutableUint8) Value() uint8 {
	return Uint8FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableUint8 struct {
	ScImmutableUint8
}

func NewScMutableUint8(proxy Proxy) ScMutableUint8 {
	return ScMutableUint8{ScImmutableUint8{proxy: proxy}}
}

func (o ScMutableUint8) Delete() {
	o.proxy.Delete()
}

func (o ScMutableUint8) SetValue(value uint8) {
	o.proxy.Set(BytesFromUint8(value))
}
