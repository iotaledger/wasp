// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ScTokenIDLength = 38

type ScTokenID struct {
	id [ScTokenIDLength]byte
}

func (o ScTokenID) Bytes() []byte {
	return TokenIDToBytes(o)
}

func (o ScTokenID) String() string {
	return TokenIDToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func TokenIDDecode(dec *WasmDecoder) ScTokenID {
	return tokenIDFromBytesUnchecked(dec.FixedBytes(ScTokenIDLength))
}

func TokenIDEncode(enc *WasmEncoder, value ScTokenID) {
	enc.FixedBytes(value.id[:], ScTokenIDLength)
}

func TokenIDFromBytes(buf []byte) ScTokenID {
	if len(buf) == 0 {
		return ScTokenID{}
	}
	if len(buf) != ScTokenIDLength {
		panic("invalid TokenID length")
	}
	return tokenIDFromBytesUnchecked(buf)
}

func TokenIDToBytes(value ScTokenID) []byte {
	return value.id[:]
}

func TokenIDFromString(value string) ScTokenID {
	return TokenIDFromBytes(HexDecode(value[2:]))
}

func TokenIDToString(value ScTokenID) string {
	return "0x" + HexEncode(TokenIDToBytes(value))
}

func tokenIDFromBytesUnchecked(buf []byte) ScTokenID {
	o := ScTokenID{}
	copy(o.id[:], buf)
	return o
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableTokenID struct {
	proxy Proxy
}

func NewScImmutableTokenID(proxy Proxy) ScImmutableTokenID {
	return ScImmutableTokenID{proxy: proxy}
}

func (o ScImmutableTokenID) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableTokenID) String() string {
	return TokenIDToString(o.Value())
}

func (o ScImmutableTokenID) Value() ScTokenID {
	return TokenIDFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableTokenID struct {
	ScImmutableTokenID
}

func NewScMutableTokenID(proxy Proxy) ScMutableTokenID {
	return ScMutableTokenID{ScImmutableTokenID{proxy: proxy}}
}

func (o ScMutableTokenID) Delete() {
	o.proxy.Delete()
}

func (o ScMutableTokenID) SetValue(value ScTokenID) {
	o.proxy.Set(TokenIDToBytes(value))
}
