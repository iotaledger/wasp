// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

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

func (o ScColor) Bytes() []byte {
	return ColorToBytes(o)
}

func (o ScColor) String() string {
	return ColorToString(o)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func ColorDecode(dec *WasmDecoder) ScColor {
	return colorFromBytesUnchecked(dec.FixedBytes(ScColorLength))
}

func ColorEncode(enc *WasmEncoder, value ScColor) {
	enc.FixedBytes(value.Bytes(), ScColorLength)
}

func ColorFromBytes(buf []byte) ScColor {
	if len(buf) == 0 {
		return ScColor{}
	}
	if len(buf) != ScColorLength {
		panic("invalid Color length")
	}
	return colorFromBytesUnchecked(buf)
}

func ColorToBytes(value ScColor) []byte {
	return value.id[:]
}

func ColorToString(value ScColor) string {
	// TODO standardize human readable string
	return Base58Encode(value.id[:])
}

func colorFromBytesUnchecked(buf []byte) ScColor {
	o := ScColor{}
	copy(o.id[:], buf)
	return o
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
	return ColorToString(o.Value())
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
	o.proxy.Set(ColorToBytes(value))
}
