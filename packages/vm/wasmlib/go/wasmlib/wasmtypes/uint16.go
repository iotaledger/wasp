// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package wasmtypes

import (
	"encoding/binary"
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

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
	if len(buf) != 2 {
		Panic("invalid Uint16 length")
	}
	return binary.LittleEndian.Uint16(buf)
}

func BytesFromUint16(value uint16) []byte {
	tmp := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmp, value)
	return tmp
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
