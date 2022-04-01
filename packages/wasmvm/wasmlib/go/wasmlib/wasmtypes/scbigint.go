// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

type ScBigInt struct {
	bytes []byte
}

func NewScBigInt(value ...uint64) ScBigInt {
	if len(value) == 0 {
		return ScBigInt{}
	}
	o := ScBigInt{}
	o.SetUint64(value[0])
	return o
}

func (o ScBigInt) Add(rhs ScBigInt) ScBigInt {
	lhsLen := len(o.bytes)
	rhsLen := len(rhs.bytes)
	if lhsLen < rhsLen {
		// always add shorter value to longer value
		return rhs.Add(o)
	}

	carry := uint16(0)
	for i := 0; i < rhsLen; i++ {
		carry += uint16(o.bytes[i]) + uint16(rhs.bytes[i])
		o.bytes[i] = byte(carry)
		carry >>= 8
	}
	for i := rhsLen; carry != 0 && i < lhsLen; i++ {
		carry += uint16(o.bytes[i])
		o.bytes[i] = byte(carry)
		carry >>= 8
	}
	if carry != 0 {
		o.bytes = append(o.bytes, 1)
	}
	return o
}

func (o ScBigInt) Bytes() []byte {
	return BigIntToBytes(o)
}

func (o ScBigInt) Cmp(rhs ScBigInt) int {
	lhsLen := len(o.bytes)
	rhsLen := len(rhs.bytes)
	if lhsLen != rhsLen {
		if lhsLen > rhsLen {
			return 1
		}
		return -1
	}
	for i := lhsLen - 1; i >= 0; i-- {
		lhsByte := o.bytes[i]
		rhsByte := rhs.bytes[i]
		if lhsByte != rhsByte {
			if lhsByte > rhsByte {
				return 1
			}
			return -1
		}
	}
	return 0
}

func (o ScBigInt) Div(rhs ScBigInt) ScBigInt {
	div, _ := o.DivMod(rhs)
	return div
}

func (o ScBigInt) DivMod(rhs ScBigInt) (ScBigInt, ScBigInt) {
	panic("implement DivMod")
	return o, rhs
}

func (o ScBigInt) IsUint64() bool {
	return len(o.bytes) <= ScUint64Length
}

func (o ScBigInt) IsZero() bool {
	return len(o.bytes) == 0
}

func (o ScBigInt) Modulo(rhs ScBigInt) ScBigInt {
	_, mod := o.DivMod(rhs)
	return mod
}

func (o ScBigInt) Mul(rhs ScBigInt) ScBigInt {
	lhsLen := len(o.bytes)
	rhsLen := len(rhs.bytes)
	if lhsLen < rhsLen {
		// always multiply bigger value by smaller value
		return rhs.Mul(o)
	}
	panic("implement Mul")
	return o
}

func (o *ScBigInt) normalize() {
	buf := o.bytes
	bufLen := len(buf)
	for ; bufLen > 0 && buf[bufLen-1] == 0; bufLen-- {
	}
	o.bytes = buf[:bufLen]
}

func (o *ScBigInt) SetUint64(value uint64) {
	o.bytes = Uint64ToBytes(value)
	o.normalize()
}

func (o ScBigInt) String() string {
	return BigIntToString(o)
}

func (o ScBigInt) Sub(rhs ScBigInt) ScBigInt {
	cmp := o.Cmp(rhs)
	if cmp <= 0 {
		if cmp < 0 {
			panic("subtraction underflow")
		}
		return ScBigInt{}
	}
	lhsLen := len(o.bytes)
	rhsLen := len(rhs.bytes)

	borrow := uint16(0)
	for i := 0; i < rhsLen; i++ {
		borrow += uint16(o.bytes[i]) - uint16(rhs.bytes[i])
		o.bytes[i] = byte(borrow)
		borrow >>= 8
	}
	for i := rhsLen; borrow != 0 && i < lhsLen; i++ {
		borrow += uint16(o.bytes[i])
		o.bytes[i] = byte(borrow)
		borrow >>= 8
	}
	o.normalize()
	return o
}

func (o ScBigInt) Uint64() uint64 {
	if len(o.bytes) > ScUint64Length {
		panic("value exceeds Uint64")
	}
	buf := make([]byte, ScUint64Length)
	copy(buf, o.bytes)
	return Uint64FromBytes(buf)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

func BigIntDecode(dec *WasmDecoder) ScBigInt {
	return ScBigInt{bytes: dec.Bytes()}
}

func BigIntEncode(enc *WasmEncoder, value ScBigInt) {
	enc.Bytes(value.bytes)
}

func BigIntFromBytes(buf []byte) ScBigInt {
	return ScBigInt{bytes: buf}
}

func BigIntToBytes(value ScBigInt) []byte {
	return value.bytes
}

func BigIntToString(value ScBigInt) string {
	// TODO standardize human readable string
	panic("implement BigIntToString")
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

func (o ScImmutableBigInt) Value() ScBigInt {
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

func (o ScMutableBigInt) SetValue(value ScBigInt) {
	o.proxy.Set(BigIntToBytes(value))
}
