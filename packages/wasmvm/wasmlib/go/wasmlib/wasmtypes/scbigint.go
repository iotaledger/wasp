// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package wasmtypes

import (
	"math/big"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func BigIntDecode(dec *WasmDecoder) *big.Int {
	value := new(big.Int)
	value.SetBytes(dec.Bytes())
	return value
}

func BigIntEncode(enc *WasmEncoder, value *big.Int) {
	enc.Bytes(value.Bytes())
}

func BigIntFromBytes(buf []byte) *big.Int {
	value := new(big.Int)
	if len(buf) == 0 {
		return value
	}
	return value.SetBytes(buf)
}

func BigIntToBytes(value *big.Int) []byte {
	return value.Bytes()
}

func BigIntToString(value *big.Int) string {
	return value.String()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableBigInt struct {
	proxy Proxy
}

func NewScImmutableBigInt(proxy Proxy) ScImmutableBigInt {
	return ScImmutableBigInt{proxy: proxy}
}

func (o ScImmutableBigInt) Exists() bool {
	return o.proxy.Exists()
}

func (o ScImmutableBigInt) String() string {
	return BigIntToString(o.Value())
}

func (o ScImmutableBigInt) Value() *big.Int {
	return BigIntFromBytes(o.proxy.Get())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableBigInt struct {
	ScImmutableBigInt
}

func NewScMutableBigInt(proxy Proxy) ScMutableBigInt {
	return ScMutableBigInt{ScImmutableBigInt{proxy: proxy}}
}

func (o ScMutableBigInt) Delete() {
	o.proxy.Delete()
}

func (o ScMutableBigInt) SetValue(value *big.Int) {
	o.proxy.Set(BigIntToBytes(value))
}
