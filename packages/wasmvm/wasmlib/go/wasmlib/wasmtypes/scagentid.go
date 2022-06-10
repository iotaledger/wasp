// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"strings"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const (
	ScAgentIDNil      byte = 0
	ScAgentIDAddress  byte = 1
	ScAgentIDContract byte = 2
	ScAgentIDEthereum byte = 3
)

type ScAgentID struct {
	kind    byte
	address ScAddress
	hname   ScHname
}

const nilAgentIDString = "-"

func NewScAgentID(address ScAddress, hname ScHname) ScAgentID {
	return ScAgentID{kind: ScAgentIDContract, address: address, hname: hname}
}

func NewScAgentIDFromAddress(address ScAddress) ScAgentID {
	if address.id[0] == ScAddressAlias {
		return NewScAgentID(address, 0)
	}
	switch address.id[0] {
	case ScAddressAlias:
		return NewScAgentID(address, 0)
	case ScAddressEth:
		return ScAgentID{kind: ScAgentIDEthereum, address: address, hname: 0}
	default:
		return ScAgentID{kind: ScAgentIDAddress, address: address, hname: 0}
	}
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
	return o.kind == ScAgentIDAddress
}

func (o ScAgentID) IsContract() bool {
	return o.kind == ScAgentIDContract
}

func (o ScAgentID) String() string {
	return AgentIDToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func AgentIDDecode(dec *WasmDecoder) ScAgentID {
	return AgentIDFromBytes(dec.Bytes())
}

func AgentIDEncode(enc *WasmEncoder, value ScAgentID) {
	enc.Bytes(AgentIDToBytes(value))
}

func AgentIDFromBytes(buf []byte) (a ScAgentID) {
	if len(buf) == 0 {
		return a
	}
	a.kind = buf[0]
	buf = buf[1:]
	switch a.kind {
	case ScAgentIDAddress:
		if len(buf) != ScLengthAlias && len(buf) != ScLengthEd25519 {
			panic("invalid AgentID length: address agentID")
		}
		a.address = AddressFromBytes(buf)
	case ScAgentIDContract:
		if len(buf) != ScChainIDLength+ScHnameLength {
			panic("invalid AgentID length: contract agentID")
		}
		a.address = ChainIDFromBytes(buf[:ScChainIDLength]).Address()
		a.hname = HnameFromBytes(buf[ScChainIDLength:])
	case ScAgentIDEthereum:
		if len(buf) != ScLengthETH {
			panic("invalid AgentID length: eth agentID")
		}
		a.address = AddressFromBytes(buf)
	case ScAgentIDNil:
		break
	default:
		panic("AgentIDFromBytes: invalid AgentID type")
	}
	return a
}

func AgentIDToBytes(value ScAgentID) []byte {
	buf := []byte{value.kind}
	switch value.kind {
	case ScAgentIDAddress:
		return append(buf, AddressToBytes(value.address)...)
	case ScAgentIDContract:
		buf = append(buf, AddressToBytes(value.address)[1:]...)
		return append(buf, HnameToBytes(value.hname)...)
	case ScAgentIDEthereum:
		return append(buf, AddressToBytes(value.address)...)
	case ScAgentIDNil:
		return buf
	default:
		panic("AgentIDToBytes: invalid AgentID type")
	}
}

func AgentIDFromString(value string) ScAgentID {
	if value == nilAgentIDString {
		return ScAgentID{}
	}

	parts := strings.Split(value, "@")
	switch len(parts) {
	case 1:
		return NewScAgentIDFromAddress(AddressFromString(parts[0]))
	case 2:
		return NewScAgentID(AddressFromString(parts[1]), HnameFromString(parts[0]))
	default:
		panic("invalid AgentID string")
	}
}

func AgentIDToString(value ScAgentID) string {
	switch value.kind {
	case ScAgentIDAddress:
		return AddressToString(value.Address())
	case ScAgentIDContract:
		return HnameToString(value.Hname()) + "@" + AddressToString(value.Address())
	case ScAgentIDEthereum:
		return AddressToString(value.Address())
	case ScAgentIDNil:
		// iscp.NilAgentID.String() returns "-" which means NilAgentID is "-"
		return nilAgentIDString
	default:
		panic("AgentIDToString: invalid AgentID type")
	}
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
