// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScAgentIDLength = ScAddressLength + ScHnameLength

type ScAgentID struct {
	address ScAddress
	hname   ScHname
}

func DecodeAgentID(dec *wasmcodec.WasmDecoder) ScAgentID {
	return newAgentIDFromBytes(dec.FixedBytes(ScAgentIDLength))
}

func EncodeAgentID(enc *wasmcodec.WasmEncoder, value ScAgentID) {
	EncodeAddress(enc, value.address)
	EncodeHname(enc, value.hname)
}

func AgentIDFromBytes(buf []byte) ScAgentID {
	if buf == nil {
		return ScAgentID{}
	}
	if len(buf) != ScAgentIDLength {
		Panic("invalid AgentID length")
	}
	// max ledgerstate.AliasAddressType
	if buf[0] > 2 {
		Panic("invalid AgentID: address type > 2")
	}
	return newAgentIDFromBytes(buf)
}

func newAgentIDFromBytes(buf []byte) ScAgentID {
	return ScAgentID{
		address: AddressFromBytes(buf[:ScAddressLength]),
		hname:   HnameFromBytes(buf[ScAddressLength:]),
	}
}

func NewScAgentID(address ScAddress, hname ScHname) ScAgentID {
	return ScAgentID{address: address, hname: hname}
}

func (o ScAgentID) Address() ScAddress {
	return o.address
}

func (o ScAgentID) Bytes() []byte {
	enc := wasmcodec.NewWasmEncoder()
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
