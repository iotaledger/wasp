// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"encoding/binary"
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScUint32Length = 4

func DecodeUint32(dec *wasmcodec.WasmDecoder) uint32 {
	return uint32(dec.VluDecode(32))
}

func EncodeUint32(enc *wasmcodec.WasmEncoder, value uint32) {
	enc.VluEncode(uint64(value))
}

func Uint32FromBytes(buf []byte) uint32 {
	if buf == nil {
		return 0
	}
	if len(buf) != ScUint32Length {
		Panic("invalid Uint32 length")
	}
	return binary.LittleEndian.Uint32(buf)
}

func BytesFromUint32(value uint32) []byte {
	tmp := make([]byte, ScUint32Length)
	binary.LittleEndian.PutUint32(tmp, value)
	return tmp
}

func StringFromUint32(value uint32) string {
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
	return StringFromUint32(o.Value())
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
	o.proxy.Set(BytesFromUint32(value))
}
