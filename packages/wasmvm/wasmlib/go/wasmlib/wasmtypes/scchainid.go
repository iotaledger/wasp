// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScChainIDLength = 20

type ScChainID struct {
	id [ScChainIDLength]byte
}

// Address returns the alias address that the chain ID actually represents
func (o ScChainID) Address() ScAddress {
	buf := []byte{ScAddressAlias}
	return AddressFromBytes(append(buf, o.id[:]...))
}

func (o ScChainID) Bytes() []byte {
	return ChainIDToBytes(o)
}

func (o ScChainID) String() string {
	return ChainIDToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func ChainIDDecode(dec *WasmDecoder) ScChainID {
	return chainIDFromBytesUnchecked(dec.FixedBytes(ScChainIDLength))
}

func ChainIDEncode(enc *WasmEncoder, value ScChainID) {
	enc.FixedBytes(value.id[:], ScChainIDLength)
}

func ChainIDFromBytes(buf []byte) ScChainID {
	if len(buf) == 0 {
		return ScChainID{}
	}
	if len(buf) != ScChainIDLength {
		panic("invalid ChainID length")
	}
	return chainIDFromBytesUnchecked(buf)
}

func ChainIDToBytes(value ScChainID) []byte {
	return value.id[:]
}

func ChainIDFromString(value string) ScChainID {
	return ChainIDFromBytes(Base58Decode(value))
}

func ChainIDToString(value ScChainID) string {
	// TODO standardize human readable string
	return Base58Encode(value.id[:])
}

func chainIDFromBytesUnchecked(buf []byte) ScChainID {
	o := ScChainID{}
	copy(o.id[:], buf)
	return o
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableChainID struct {
	proxy Proxy
}

func NewScImmutableChainID(proxy Proxy) ScImmutableChainID {
	return ScImmutableChainID{proxy: proxy}
}

func (o ScImmutableChainID) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableChainID) String() string {
	return ChainIDToString(o.Value())
}

func (o ScImmutableChainID) Value() ScChainID {
	return ChainIDFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableChainID struct {
	ScImmutableChainID
}

func NewScMutableChainID(proxy Proxy) ScMutableChainID {
	return ScMutableChainID{ScImmutableChainID{proxy: proxy}}
}

func (o ScMutableChainID) Delete() {
	o.proxy.Delete()
}

func (o ScMutableChainID) SetValue(value ScChainID) {
	o.proxy.Set(ChainIDToBytes(value))
}
