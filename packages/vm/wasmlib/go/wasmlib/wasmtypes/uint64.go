// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"encoding/binary"
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeUint64(dec *wasmcodec.WasmDecoder) uint64 {
	return dec.VluDecode(64)
}

func EncodeUint64(enc *wasmcodec.WasmEncoder, value uint64) {
	enc.VluEncode(value)
}

func Uint64FromBytes(buf []byte) uint64 {
	if buf == nil {
		return 0
	}
	if len(buf) != 8 {
		Panic("invalid Uint64 length")
	}
	return binary.LittleEndian.Uint64(buf)
}

func BytesFromUint64(value uint64) []byte {
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, value)
	return tmp
}

func StringFromUint64(value uint64) string {
	return strconv.FormatUint(value, 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableUint64 struct {
	proxy Proxy
}

func NewScImmutableUint64(proxy Proxy) ScImmutableUint64 {
	return ScImmutableUint64{proxy: proxy}
}

func (o ScImmutableUint64) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableUint64) String() string {
	return StringFromUint64(o.Value())
}

func (o ScImmutableUint64) Value() uint64 {
	return Uint64FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableUint64 struct {
	ScImmutableUint64
}

func NewScMutableUint64(proxy Proxy) ScMutableUint64 {
	return ScMutableUint64{ScImmutableUint64{proxy: proxy}}
}

func (o ScMutableUint64) Delete() {
	o.proxy.Delete()
}

func (o ScMutableUint64) SetValue(value uint64) {
	o.proxy.Set(BytesFromUint64(value))
}
