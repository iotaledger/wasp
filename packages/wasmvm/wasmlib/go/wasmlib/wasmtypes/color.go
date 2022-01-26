// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

const ScColorLength = 32

type ScColor struct {
	id [ScColorLength]byte
}

var (
	IOTA = ScColor{}
	MINT = ScColor{}
)

func init() {
	for i := range MINT.id {
		MINT.id[i] = 0xff
	}
}

func newColorFromBytes(buf []byte) ScColor {
	o := ScColor{}
	copy(o.id[:], buf)
	return o
}

func (o ScColor) Bytes() []byte {
	return o.id[:]
}

func (o ScColor) String() string {
	// TODO standardize human readable string
	return base58Encode(o.id[:])
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func DecodeColor(dec *WasmDecoder) ScColor {
	return newColorFromBytes(dec.FixedBytes(ScColorLength))
}

func EncodeColor(enc *WasmEncoder, value ScColor) {
	enc.FixedBytes(value.Bytes(), ScColorLength)
}

func ColorFromBytes(buf []byte) ScColor {
	if buf == nil {
		return ScColor{}
	}
	if len(buf) != ScColorLength {
		panic("invalid Color length")
	}
	return newColorFromBytes(buf)
}

func BytesFromColor(value ScColor) []byte {
	return value.Bytes()
}

func ColorToString(value ScColor) string {
	return value.String()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableColor struct {
	proxy Proxy
}

func NewScImmutableColor(proxy Proxy) ScImmutableColor {
	return ScImmutableColor{proxy: proxy}
}

func (o ScImmutableColor) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableColor) String() string {
	return o.Value().String()
}

func (o ScImmutableColor) Value() ScColor {
	return ColorFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableColor struct {
	ScImmutableColor
}

func NewScMutableColor(proxy Proxy) ScMutableColor {
	return ScMutableColor{ScImmutableColor{proxy: proxy}}
}

func (o ScMutableColor) Delete() {
	o.proxy.Delete()
}

func (o ScMutableColor) SetValue(value ScColor) {
	o.proxy.Set(value.Bytes())
}
