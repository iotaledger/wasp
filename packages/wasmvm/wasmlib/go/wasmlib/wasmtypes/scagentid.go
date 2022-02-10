// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const nilAgentID = 0xff

type ScAgentID struct {
	address ScAddress
	hname   ScHname
}

var nilAgent = ScAgentID{}

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
	if dec.Peek() == ScAddressEd25519 {
		return ScAgentID{address: AddressDecode(dec)}
	}
	return ScAgentID{address: AddressDecode(dec), hname: HnameDecode(dec)}
}

func AgentIDEncode(enc *WasmEncoder, value ScAgentID) {
	AddressEncode(enc, value.address)
	if value.address.Bytes()[0] != ScAddressEd25519 {
		HnameEncode(enc, value.hname)
	}
}

func AgentIDFromBytes(buf []byte) ScAgentID {
	if len(buf) == 0 {
		return ScAgentID{}
	}
	switch buf[0] {
	case ScAddressAlias:
		if len(buf) != ScLengthAlias+ScHnameLength {
			panic("invalid AgentID length: Alias address")
		}
		return ScAgentID{
			address: AddressFromBytes(buf[:ScLengthAlias]),
			hname:   HnameFromBytes(buf[ScLengthAlias:]),
		}
	case ScAddressEd25519:
		if len(buf) != ScLengthEd25519 {
			panic("invalid AgentID length: Ed25519 address")
		}
		return ScAgentID{
			address: AddressFromBytes(buf),
		}
	case ScAddressNFT:
		if len(buf) != ScLengthNFT+ScHnameLength {
			panic("invalid AgentID length: NFT address")
		}
		return ScAgentID{
			address: AddressFromBytes(buf[:ScLengthNFT]),
			hname:   HnameFromBytes(buf[ScLengthNFT:]),
		}
	case nilAgentID: // nil agent id
		if len(buf) != 1 {
			panic("invalid AgentID length: nil AgentID")
		}
		return ScAgentID{}
	default:
		panic("invalid AgentID address type")
	}
}

func AgentIDToBytes(value ScAgentID) []byte {
	if value == nilAgent {
		return []byte{nilAgentID}
	}
	buf := AddressToBytes(value.address)
	if buf[0] == ScAddressEd25519 {
		return buf
	}
	return append(buf, HnameToBytes(value.hname)...)
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
