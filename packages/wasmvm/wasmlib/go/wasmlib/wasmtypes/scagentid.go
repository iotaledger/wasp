// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScAgentIDLength = ScAddressLength + ScHnameLength

type ScAgentID struct {
	address ScAddress
	hname   ScHname
}

func NewScAgentID(address ScAddress, hname ScHname) ScAgentID {
	return ScAgentID{address: address, hname: hname}
}

func (o ScAgentID) Address() ScAddress {
	return o.address
}

func (o ScAgentID) Bytes() []byte {
	return AgentIDToBytes(o)
}

func (o ScAgentID) Hname() ScHname {
	return o.hname
}

func (o ScAgentID) IsAddress() bool {
	return o.hname == 0
}

func (o ScAgentID) String() string {
	return AgentIDToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func AgentIDDecode(dec *WasmDecoder) ScAgentID {
	return ScAgentID{
		address: AddressDecode(dec),
		hname:   HnameDecode(dec),
	}
}

func AgentIDEncode(enc *WasmEncoder, value ScAgentID) {
	AddressEncode(enc, value.address)
	HnameEncode(enc, value.hname)
}

func AgentIDFromBytes(buf []byte) ScAgentID {
	if len(buf) == 0 {
		return ScAgentID{}
	}
	if len(buf) != ScAgentIDLength {
		panic("invalid AgentID length")
	}
	if buf[0] > AddressNFT {
		panic("invalid AgentID address type")
	}
	return ScAgentID{
		address: AddressFromBytes(buf[:ScAddressLength]),
		hname:   HnameFromBytes(buf[ScAddressLength:]),
	}
}

func AgentIDToBytes(value ScAgentID) []byte {
	enc := NewWasmEncoder()
	AgentIDEncode(enc, value)
	return enc.Buf()
}

func AgentIDToString(value ScAgentID) string {
	// TODO standardize human readable string
	return value.address.String() + "::" + value.hname.String()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableAgentID struct {
	proxy Proxy
}

func NewScImmutableAgentID(proxy Proxy) ScImmutableAgentID {
	return ScImmutableAgentID{proxy: proxy}
}

func (o ScImmutableAgentID) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableAgentID) String() string {
	return AgentIDToString(o.Value())
}

func (o ScImmutableAgentID) Value() ScAgentID {
	return AgentIDFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableAgentID struct {
	ScImmutableAgentID
}

func NewScMutableAgentID(proxy Proxy) ScMutableAgentID {
	return ScMutableAgentID{ScImmutableAgentID{proxy: proxy}}
}

func (o ScMutableAgentID) Delete() {
	o.proxy.Delete()
}

func (o ScMutableAgentID) SetValue(value ScAgentID) {
	o.proxy.Set(AgentIDToBytes(value))
}
