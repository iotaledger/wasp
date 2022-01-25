// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"encoding/binary"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScHnameLength = 4

type ScHname uint32

func DecodeHname(dec *wasmcodec.WasmDecoder) ScHname {
	return newHnameFromBytes(dec.FixedBytes(ScHnameLength))
}

func EncodeHname(enc *wasmcodec.WasmEncoder, value ScHname) {
	enc.FixedBytes(value.Bytes(), ScHnameLength)
}

func HnameFromBytes(buf []byte) ScHname {
	if buf == nil {
		return 0
	}
	if len(buf) != ScHnameLength {
		Panic("invalid Hname length")
	}
	return newHnameFromBytes(buf)
}

func newHnameFromBytes(buf []byte) ScHname {
	return ScHname(binary.LittleEndian.Uint32(buf))
}

//func HnameFromString(name string) ScHname {
//	return HnameFromBytes(wasmlib.Sandbox(wasmstore.FnUtilsHashName, []byte(name)))
//}

func (o ScHname) Bytes() []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(o))
	return buf
}

func (o ScHname) String() string {
	// TODO standardize human readable string
	return hex(o.Bytes())
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
	return o.Value().String()
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
	o.proxy.Set(value.Bytes())
}
