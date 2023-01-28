// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"strconv"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScHnameLength = 4

type ScHname uint32

func NewScHname(name string) ScHname {
	return HashName(name)
}

func (o ScHname) Bytes() []byte {
	return HnameToBytes(o)
}

func (o ScHname) String() string {
	return HnameToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func HnameDecode(dec *WasmDecoder) ScHname {
	return hnameFromBytesUnchecked(dec.FixedBytes(ScHnameLength))
}

func HnameEncode(enc *WasmEncoder, value ScHname) {
	enc.FixedBytes(value.Bytes(), ScHnameLength)
}

func HnameFromBytes(buf []byte) ScHname {
	if len(buf) == 0 {
		return 0
	}
	if len(buf) != ScHnameLength {
		panic("invalid Hname length")
	}
	return hnameFromBytesUnchecked(buf)
}

func HnameToBytes(value ScHname) []byte {
	return Uint32ToBytes(uint32(value))
}

func HnameFromString(value string) ScHname {
	if len(value) > 8 {
		panic("invalid Hname string")
	}
	u64, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		panic(err)
	}
	return ScHname(u64)
}

func HnameToString(value ScHname) string {
	res := strconv.FormatUint(uint64(value), 16)
	return "0000000"[:8-len(res)] + res
}

func hnameFromBytesUnchecked(buf []byte) ScHname {
	return ScHname(Uint32FromBytes(buf))
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableHname struct {
	proxy Proxy
}

func NewScImmutableHname(proxy Proxy) ScImmutableHname {
	return ScImmutableHname{proxy: proxy}
}

func (o ScImmutableHname) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableHname) String() string {
	return HnameToString(o.Value())
}

func (o ScImmutableHname) Value() ScHname {
	return HnameFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableHname struct {
	ScImmutableHname
}

func NewScMutableHname(proxy Proxy) ScMutableHname {
	return ScMutableHname{ScImmutableHname{proxy: proxy}}
}

func (o ScMutableHname) Delete() {
	o.proxy.Delete()
}

func (o ScMutableHname) SetValue(value ScHname) {
	o.proxy.Set(HnameToBytes(value))
}
