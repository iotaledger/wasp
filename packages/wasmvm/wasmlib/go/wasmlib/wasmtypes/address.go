// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

const ScAddressLength = 33

type ScAddress struct {
	id [ScAddressLength]byte
}

func addressFromBytes(buf []byte) ScAddress {
	o := ScAddress{}
	copy(o.id[:], buf)
	return o
}

func (o ScAddress) AsAgentID() ScAgentID {
	// agentID for address has Hname zero
	return NewScAgentID(o, 0)
}

func (o ScAddress) Bytes() []byte {
	return o.id[:]
}

func (o ScAddress) String() string {
	// TODO standardize human readable string
	return base58Encode(o.id[:])
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeAddress(dec *WasmDecoder) ScAddress {
	return addressFromBytes(dec.FixedBytes(ScAddressLength))
}

func EncodeAddress(enc *WasmEncoder, value ScAddress) {
	enc.FixedBytes(value.Bytes(), ScAddressLength)
}

func AddressFromBytes(buf []byte) ScAddress {
	if buf == nil {
		return ScAddress{}
	}
	if len(buf) != ScAddressLength {
		panic("invalid Address length")
	}
	// max ledgerstate.AliasAddressType
	if buf[0] > 2 {
		panic("invalid Address: address type > 2")
	}
	return addressFromBytes(buf)
}

func BytesFromAddress(value ScAddress) []byte {
	return value.Bytes()
}

func AddressToString(value ScAddress) string {
	return value.String()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableAddress struct {
	proxy Proxy
}

func NewScImmutableAddress(proxy Proxy) ScImmutableAddress {
	return ScImmutableAddress{proxy: proxy}
}

func (o ScImmutableAddress) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableAddress) String() string {
	return o.Value().String()
}

func (o ScImmutableAddress) Value() ScAddress {
	return AddressFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableAddress struct {
	ScImmutableAddress
}

func NewScMutableAddress(proxy Proxy) ScMutableAddress {
	return ScMutableAddress{ScImmutableAddress{proxy: proxy}}
}

func (o ScMutableAddress) Delete() {
	o.proxy.Delete()
}

func (o ScMutableAddress) SetValue(value ScAddress) {
	o.proxy.Set(value.Bytes())
}
