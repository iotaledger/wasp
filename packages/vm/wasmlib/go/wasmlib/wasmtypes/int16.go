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

func DecodeInt16(dec *wasmcodec.WasmDecoder) int16 {
	return int16(dec.VliDecode(16))
}

func EncodeInt16(enc *wasmcodec.WasmEncoder, value int16) {
	enc.VliEncode(int64(value))
}

func Int16FromBytes(buf []byte) int16 {
	if buf == nil {
		return 0
	}
	if len(buf) != 2 {
		Panic("invalid Int16 length")
	}
	return int16(binary.LittleEndian.Uint16(buf))
}

func BytesFromInt16(value int16) []byte {
	tmp := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmp, uint16(value))
	return tmp
}

func StringFromInt16(value int16) string {
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
	return StringFromInt16(o.Value())
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
	o.proxy.Set(BytesFromInt16(value))
}
