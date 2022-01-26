// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

const ScAgentIDLength = ScAddressLength + ScHnameLength

type ScAgentID struct {
	address ScAddress
	hname   ScHname
}

func NewScAgentID(address ScAddress, hname ScHname) ScAgentID {
	return ScAgentID{address: address, hname: hname}
}

func agentIDFromBytes(buf []byte) ScAgentID {
	return ScAgentID{
		address: AddressFromBytes(buf[:ScAddressLength]),
		hname:   HnameFromBytes(buf[ScAddressLength:]),
	}
}

func (o ScAgentID) Address() ScAddress {
	return o.address
}

func (o ScAgentID) Bytes() []byte {
	enc := NewWasmEncoder()
	EncodeAgentID(enc, o)
	return enc.Buf()
}

func (o ScAgentID) Hname() ScHname {
	return o.hname
}

func (o ScAgentID) IsAddress() bool {
	return o.hname == 0
}

func (o ScAgentID) String() string {
	// TODO standardize human readable string
	return o.address.String() + "::" + o.hname.String()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeAgentID(dec *WasmDecoder) ScAgentID {
	return agentIDFromBytes(dec.FixedBytes(ScAgentIDLength))
}

func EncodeAgentID(enc *WasmEncoder, value ScAgentID) {
	EncodeAddress(enc, value.address)
	EncodeHname(enc, value.hname)
}

func AgentIDFromBytes(buf []byte) ScAgentID {
	if buf == nil {
		return ScAgentID{}
	}
	if len(buf) != ScAgentIDLength {
		panic("invalid AgentID length")
	}
	// max ledgerstate.AliasAddressType
	if buf[0] > 2 {
		panic("invalid AgentID: address type > 2")
	}
	return agentIDFromBytes(buf)
}

func BytesFromAgentID(value ScAgentID) []byte {
	return value.Bytes()
}

func AgentIDToString(value ScAgentID) string {
	return value.String()
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
	return o.Value().String()
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
	o.proxy.Set(value.Bytes())
}
