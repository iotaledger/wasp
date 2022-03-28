// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScNftIDLength = 20

type ScNftID struct {
	id [ScNftIDLength]byte
}

func (o ScNftID) Bytes() []byte {
	return NftIDToBytes(&o)
}

func (o ScNftID) String() string {
	return NftIDToString(&o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func NftIDDecode(dec *WasmDecoder) ScNftID {
	return nftIDFromBytesUnchecked(dec.FixedBytes(ScNftIDLength))
}

func NftIDEncode(enc *WasmEncoder, value *ScNftID) {
	enc.FixedBytes(value.Bytes(), ScNftIDLength)
}

func NftIDFromBytes(buf []byte) ScNftID {
	if len(buf) == 0 {
		return ScNftID{}
	}
	if len(buf) != ScNftIDLength {
		panic("invalid NftID length")
	}
	return nftIDFromBytesUnchecked(buf)
}

func NftIDToBytes(value *ScNftID) []byte {
	return value.id[:]
}

func NftIDToString(value *ScNftID) string {
	// TODO standardize human readable string
	return Base58Encode(value.id[:])
}

func nftIDFromBytesUnchecked(buf []byte) ScNftID {
	o := ScNftID{}
	copy(o.id[:], buf)
	return o
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableNftID struct {
	proxy Proxy
}

func NewScImmutableNftID(proxy Proxy) ScImmutableNftID {
	return ScImmutableNftID{proxy: proxy}
}

func (o ScImmutableNftID) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableNftID) String() string {
	value := o.Value()
	return NftIDToString(&value)
}

func (o ScImmutableNftID) Value() ScNftID {
	return NftIDFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableNftID struct {
	ScImmutableNftID
}

func NewScMutableNftID(proxy Proxy) ScMutableNftID {
	return ScMutableNftID{ScImmutableNftID{proxy: proxy}}
}

func (o ScMutableNftID) Delete() {
	o.proxy.Delete()
}

func (o ScMutableNftID) SetValue(value ScNftID) {
	o.proxy.Set(NftIDToBytes(&value))
}
