// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScAddressLength = 33

type ScAddress struct {
	id [ScAddressLength]byte
}

func DecodeAddress(dec *wasmcodec.WasmDecoder) ScAddress {
	return newAddressFromBytes(dec.FixedBytes(ScAddressLength))
}

func EncodeAddress(enc *wasmcodec.WasmEncoder, value ScAddress) {
	enc.FixedBytes(value.Bytes(), ScAddressLength)
}

func AddressFromBytes(buf []byte) ScAddress {
	if buf == nil {
		return ScAddress{}
	}
	if len(buf) != ScAddressLength {
		Panic("invalid Address length")
	}
	// max ledgerstate.AliasAddressType
	if buf[0] > 2 {
		Panic("invalid Address: address type > 2")
	}
	return newAddressFromBytes(buf)
}

func newAddressFromBytes(buf []byte) ScAddress {
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
