// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

const ScChainIDLength = 33

type ScChainID struct {
	id [ScChainIDLength]byte
}

func chainIDFromBytes(buf []byte) ScChainID {
	o := ScChainID{}
	copy(o.id[:], buf)
	return o
}

func (o ScChainID) Address() ScAddress {
	return AddressFromBytes(o.id[:])
}

func (o ScChainID) Bytes() []byte {
	return o.id[:]
}

func (o ScChainID) String() string {
	// TODO standardize human readable string
	return base58Encode(o.id[:])
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeChainID(dec *WasmDecoder) ScChainID {
	return chainIDFromBytes(dec.FixedBytes(ScChainIDLength))
}

func EncodeChainID(enc *WasmEncoder, value ScChainID) {
	enc.FixedBytes(value.Bytes(), ScChainIDLength)
}

func ChainIDFromBytes(buf []byte) ScChainID {
	if buf == nil {
		return ScChainID{id: [ScChainIDLength]byte{2}}
	}
	if len(buf) != ScChainIDLength {
		panic("invalid ChainID length")
	}
	// must be ledgerstate.AliasAddressType
	if buf[0] != 2 {
		panic("invalid ChainID: not an alias address")
	}
	return chainIDFromBytes(buf)
}

func BytesFromChainID(value ScChainID) []byte {
	return value.Bytes()
}

func ChainIDToString(value ScChainID) string {
	return value.String()
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
	return o.Value().String()
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
	o.proxy.Set(value.Bytes())
}
