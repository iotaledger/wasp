// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScHashLength = 32

type ScHash struct {
	id [ScHashLength]byte
}

func DecodeHash(dec *wasmcodec.WasmDecoder) ScHash {
	return newHashFromBytes(dec.FixedBytes(ScHashLength))
}

func EncodeHash(enc *wasmcodec.WasmEncoder, value ScHash) {
	enc.FixedBytes(value.Bytes(), ScHashLength)
}

func HashFromBytes(buf []byte) ScHash {
	if len(buf) != ScHashLength {
		Panic("invalid Hash length")
	}
	return newHashFromBytes(buf)
}

func newHashFromBytes(buf []byte) ScHash {
	o := ScHash{}
	copy(o.id[:], buf)
	return o
}

func (o ScHash) Bytes() []byte {
	return o.id[:]
}

func (o ScHash) String() string {
	// TODO standardize human readable string
	return base58Encode(o.id[:])
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableHash struct {
	proxy Proxy
}

func NewScImmutableHash(proxy Proxy) ScImmutableHash {
	return ScImmutableHash{proxy: proxy}
}

func (o ScImmutableHash) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableHash) String() string {
	return o.Value().String()
}

func (o ScImmutableHash) Value() ScHash {
	return HashFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableHash struct {
	ScImmutableHash
}

func NewScMutableHash(proxy Proxy) ScMutableHash {
	return ScMutableHash{ScImmutableHash{proxy: proxy}}
}

func (o ScMutableHash) Delete() {
	o.proxy.Delete()
}

func (o ScMutableHash) SetValue(value ScHash) {
	o.proxy.Set(value.Bytes())
}
