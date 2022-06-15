// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScHashLength = 32

type ScHash struct {
	id [ScHashLength]byte
}

func (o ScHash) Bytes() []byte {
	return HashToBytes(o)
}

func (o ScHash) String() string {
	return HashToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func HashDecode(dec *WasmDecoder) ScHash {
	return hashFromBytesUnchecked(dec.FixedBytes(ScHashLength))
}

func HashEncode(enc *WasmEncoder, value ScHash) {
	enc.FixedBytes(value.id[:], ScHashLength)
}

func HashFromBytes(buf []byte) ScHash {
	if len(buf) == 0 {
		return ScHash{}
	}
	if len(buf) != ScHashLength {
		panic("invalid Hash length")
	}
	return hashFromBytesUnchecked(buf)
}

func HashToBytes(value ScHash) []byte {
	return value.id[:]
}

func HashFromString(value string) ScHash {
	return HashFromBytes(HexDecode(value))
}

func HashToString(value ScHash) string {
	// TODO standardize human readable string
	return HexEncode(HashToBytes(value))
}

func hashFromBytesUnchecked(buf []byte) ScHash {
	o := ScHash{}
	copy(o.id[:], buf)
	return o
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
	return HashToString(o.Value())
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
	o.proxy.Set(HashToBytes(value))
}
