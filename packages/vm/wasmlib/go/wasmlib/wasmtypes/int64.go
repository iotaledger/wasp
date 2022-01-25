// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"encoding/binary"
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeInt64(dec *wasmcodec.WasmDecoder) int64 {
	return dec.VliDecode(64)
}

func EncodeInt64(enc *wasmcodec.WasmEncoder, value int64) {
	enc.VliEncode(value)
}

func Int64FromBytes(buf []byte) int64 {
	if buf == nil {
		return 0
	}
	if len(buf) != 8 {
		Panic("invalid Int64 length")
	}
	return int64(binary.LittleEndian.Uint64(buf))
}

func BytesFromInt64(value int64) []byte {
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, uint64(value))
	return tmp
}

func StringFromInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableInt64 struct {
	proxy Proxy
}

func NewScImmutableInt64(proxy Proxy) ScImmutableInt64 {
	return ScImmutableInt64{proxy: proxy}
}

func (o ScImmutableInt64) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableInt64) String() string {
	return StringFromInt64(o.Value())
}

func (o ScImmutableInt64) Value() int64 {
	return Int64FromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableInt64 struct {
	ScImmutableInt64
}

func NewScMutableInt64(proxy Proxy) ScMutableInt64 {
	return ScMutableInt64{ScImmutableInt64{proxy: proxy}}
}

func (o ScMutableInt64) Delete() {
	o.proxy.Delete()
}

func (o ScMutableInt64) SetValue(value int64) {
	o.proxy.Set(BytesFromInt64(value))
}
