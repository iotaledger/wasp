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

func DecodeInt32(dec *wasmcodec.WasmDecoder) int32 {
	return int32(dec.VliDecode(32))
}

func EncodeInt32(enc *wasmcodec.WasmEncoder, value int32) {
	enc.VliEncode(int64(value))
}

func Int32FromBytes(buf []byte) int32 {
	if len(buf) != 4 {
		Panic("invalid Int32 length")
	}
	return int32(binary.LittleEndian.Uint32(buf))
}

func BytesFromInt32(value int32) []byte {
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(value))
	return tmp
}

func StringFromInt32(value int32) string {
	return strconv.FormatInt(int64(value), 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt32 struct {
	proxy Proxy
}

func NewScImmutableInt32(proxy Proxy) ScImmutableInt32 {
	return ScImmutableInt32{proxy: proxy}
}

func (o ScImmutableInt32) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableInt32) String() string {
	return StringFromInt32(o.Value())
}

func (o ScImmutableInt32) Value() int32 {
	return Int32FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableInt32 struct {
	ScImmutableInt32
}

func NewScMutableInt32(proxy Proxy) ScMutableInt32 {
	return ScMutableInt32{ScImmutableInt32{proxy: proxy}}
}

func (o ScMutableInt32) Delete() {
	o.proxy.Delete()
}

func (o ScMutableInt32) SetValue(value int32) {
	o.proxy.Set(BytesFromInt32(value))
}
