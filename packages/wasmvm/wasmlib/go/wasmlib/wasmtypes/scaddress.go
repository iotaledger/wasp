// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const (
	ScAddressAlias   byte = 8
	ScAddressEd25519 byte = 0
	ScAddressNFT     byte = 16

	ScLengthAlias   = 21
	ScLengthEd25519 = 33
	ScLengthNFT     = 21

	ScAddressLength = ScLengthEd25519
)

type ScAddress struct {
	id [ScAddressLength]byte
}

func (o ScAddress) AsAgentID() ScAgentID {
	// agentID for address has Hname zero
	return NewScAgentID(o, 0)
}

func (o ScAddress) Bytes() []byte {
	return AddressToBytes(o)
}

func (o ScAddress) String() string {
	return AddressToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// TODO address type-dependent encoding/decoding?
func AddressDecode(dec *WasmDecoder) ScAddress {
	addr := ScAddress{}
	copy(addr.id[:], dec.FixedBytes(ScAddressLength))
	return addr
}

func AddressEncode(enc *WasmEncoder, value ScAddress) {
	enc.FixedBytes(value.id[:], ScAddressLength)
}

func AddressFromBytes(buf []byte) ScAddress {
	addr := ScAddress{}
	if len(buf) == 0 {
		return addr
	}
	switch buf[0] {
	case ScAddressAlias:
		if len(buf) != ScLengthAlias {
			panic("invalid Address length: Alias")
		}
	case ScAddressEd25519:
		if len(buf) != ScLengthEd25519 {
			panic("invalid Address length: Ed25519")
		}
	case ScAddressNFT:
		if len(buf) != ScLengthNFT {
			panic("invalid Address length: NFT")
		}
	default:
		panic("invalid Address type")
	}
	copy(addr.id[:], buf)
	return addr
}

func AddressToBytes(value ScAddress) []byte {
	switch value.id[0] {
	case ScAddressAlias:
		return value.id[:ScLengthAlias]
	case ScAddressEd25519:
		return value.id[:ScLengthEd25519]
	case ScAddressNFT:
		return value.id[:ScLengthNFT]
	default:
		panic("unexpected Address type")
	}
}

func AddressFromString(value string) ScAddress {
	return AddressFromBytes(Base58Decode(value))
}

func AddressToString(value ScAddress) string {
	// TODO standardize human readable string
	return Base58Encode(value.id[:])
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
	return AddressToString(o.Value())
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
	o.proxy.Set(AddressToBytes(value))
}
