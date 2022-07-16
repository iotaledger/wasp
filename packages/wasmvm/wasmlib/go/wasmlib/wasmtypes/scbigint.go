// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

type ScBigInt struct {
	bytes []byte
}

var (
	zero = NewScBigInt()
	one  = NewScBigInt(1)
)

func NewScBigInt(value ...uint64) ScBigInt {
	if len(value) == 0 {
		return ScBigInt{}
	}
	return normalize(Uint64ToBytes(value[0]))
}

func normalize(buf []byte) ScBigInt {
	bufLen := len(buf)
	for ; bufLen > 0 && buf[bufLen-1] == 0; bufLen-- {
	}
	return ScBigInt{bytes: buf[:bufLen]}
}

func (o ScBigInt) Add(rhs ScBigInt) ScBigInt {
	lhsLen := len(o.bytes)
	rhsLen := len(rhs.bytes)
	if lhsLen < rhsLen {
		// always add shorter value to longer value
		return rhs.Add(o)
	}

	buf := make([]byte, lhsLen)
	carry := uint16(0)
	for i := 0; i < rhsLen; i++ {
		carry += uint16(o.bytes[i]) + uint16(rhs.bytes[i])
		buf[i] = byte(carry)
		carry >>= 8
	}
	for i := rhsLen; i < lhsLen; i++ {
		carry += uint16(o.bytes[i])
		buf[i] = byte(carry)
		carry >>= 8
	}
	if carry != 0 {
		buf = append(buf, 1)
	}
	return normalize(buf)
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
	if rhs.IsZero() {
		panic("divide by zero")
	}
	cmp := o.Cmp(rhs)
	if cmp <= 0 {
		if cmp < 0 {
			// divide by larger value, quo = 0, rem = lhs
			return zero, o
		}
		// divide equal values, quo = 1, rem = 0
		return one, zero
	}
	if o.IsUint64() {
		// let standard uint64 type do the heavy lifting
		lhs64 := o.Uint64()
		rhs64 := rhs.Uint64()
		return NewScBigInt(lhs64 / rhs64), NewScBigInt(lhs64 % rhs64)
	}
	if len(rhs.bytes) == 1 {
		if rhs.bytes[0] == 1 {
			// divide by 1, quo = lhs, rem = 0
			return o, zero
		}
		return o.divModSimple(rhs.bytes[0])
	}
	return o.divModNormalize(rhs)
}

func (o ScBigInt) divModNormalize(rhs ScBigInt) (ScBigInt, ScBigInt) {
	// shift divisor MSB until the high order bit is set
	// so that we get the best guess possible when dividing by MSB

	msb := rhs.bytes[len(rhs.bytes)-1]
	if (msb & 0x80) != 0 {
		// already normalized, no shifts necessary
		return divModEstimate(o, rhs)
	}

	shift := uint32(1)
	for msb <<= 1; (msb & 0x80) == 0; msb <<= 1 {
		shift++
	}

	// shift both lhs and rhs
	quo, rem := divModEstimate(o.Shl(shift), rhs.Shl(shift))
	// shift back remainder
	return quo, rem.Shr(shift)
}

func divModEstimate(lhs, rhs ScBigInt) (ScBigInt, ScBigInt) {
	lhsLen := len(lhs.bytes)
	rhsLen := len(rhs.bytes)
	if lhsLen <= rhsLen {
		if lhs.Cmp(rhs) >= 0 {
			return one, lhs.Sub(rhs)
		}
		return zero, lhs
	}

	buf := make([]byte, lhsLen-rhsLen)
	lhs16 := Uint16FromBytes(lhs.bytes[lhsLen-2:])
	rhs16 := uint16(rhs.bytes[rhsLen-1])
	res16 := lhs16 / rhs16
	if res16 > 0xff {
		// res16 can be up to 0x0101, reduce guess to the nearest byte value
		res16 = 0xff
	}
	buf[len(buf)-1] = byte(res16)
	guess := normalize(buf)
	product := rhs.Mul(guess)

	cmp := product.Cmp(lhs)
	if cmp == 0 {
		// lucky guess and exactly divisible
		return guess, zero
	}

	if cmp < 0 {
		// underestimated, correct guess by adding estimate on remainder
		quo, rem := divModEstimate(lhs.Sub(product), rhs)
		return guess.Add(quo), rem
	}

	// overestimated, correct guess by subtracting estimate on surplus
	quo, rem := divModEstimate(product.Sub(lhs), rhs)
	if rem.IsZero() {
		return guess.Sub(quo), rem
	}
	return guess.Sub(quo).Sub(one), rhs.Sub(rem)
}

func (o ScBigInt) divModSimple(value byte) (ScBigInt, ScBigInt) {
	lhsLen := len(o.bytes)
	buf := make([]byte, lhsLen)
	remain := uint16(0)
	rhs := uint16(value)
	for i := lhsLen - 1; i >= 0; i-- {
		remain = (remain << 8) + uint16(o.bytes[i])
		buf[i] = byte(remain / rhs)
		remain %= rhs
	}
	return normalize(buf), normalize([]byte{byte(remain)})
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
	if lhsLen+rhsLen <= ScUint64Length {
		return NewScBigInt(o.Uint64() * rhs.Uint64())
	}
	if rhsLen == 0 {
		// multiply by zero, result zero
		return zero
	}
	if rhsLen == 1 && rhs.bytes[0] == 1 {
		// multiply by one, result lhs
		return o
	}

	// TODO optimize by using u32 words instead of u8 words
	buf := make([]byte, lhsLen+rhsLen)
	for r := 0; r < rhsLen; r++ {
		carry := uint16(0)
		for l := 0; l < lhsLen; l++ {
			carry += uint16(buf[l+r]) + uint16(o.bytes[l])*uint16(rhs.bytes[r])
			buf[l+r] = byte(carry)
			carry >>= 8
		}
		buf[r+lhsLen] = byte(carry)
	}
	return normalize(buf)
}

func (o ScBigInt) Shl(shift uint32) ScBigInt {
	if shift == 0 {
		return o
	}

	wholeBytes := int(shift >> 3)
	shift &= 0x07

	lhsLen := len(o.bytes)
	bufLen := lhsLen + wholeBytes + 1
	buf := make([]byte, bufLen)
	word := uint16(0)
	for i := lhsLen; i > 0; i-- {
		word = (word << 8) + uint16(o.bytes[i-1])
		bufLen--
		buf[bufLen] = byte(word >> (8 - shift))
	}
	buf[bufLen-1] = byte(word << shift)
	return normalize(buf)
}

func (o ScBigInt) Shr(shift uint32) ScBigInt {
	if shift == 0 {
		return o
	}

	wholeBytes := int(shift >> 3)
	shift &= 0x07

	lhsLen := len(o.bytes)
	if wholeBytes >= lhsLen {
		return zero
	}

	bufLen := lhsLen - wholeBytes
	buf := make([]byte, bufLen)
	bytes := o.bytes[wholeBytes:]
	word := uint16(bytes[0]) << 8
	for i := 1; i < bufLen; i++ {
		word = (word >> 8) + (uint16(bytes[i]) << 8)
		buf[i-1] = byte(word >> shift)
	}
	buf[bufLen-1] = byte(word >> (8 + shift))
	return normalize(buf)
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

	buf := make([]byte, lhsLen)
	borrow := uint16(0)
	for i := 0; i < rhsLen; i++ {
		borrow += uint16(o.bytes[i]) - uint16(rhs.bytes[i])
		buf[i] = byte(borrow)
		borrow = uint16(int16(borrow) >> 8)
	}
	for i := rhsLen; i < lhsLen; i++ {
		borrow += uint16(o.bytes[i])
		buf[i] = byte(borrow)
		borrow = uint16(int16(borrow) >> 8)
	}
	return normalize(buf)
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

var quintillion = NewScBigInt(1_000_000_000_000_000_000)

func BigIntDecode(dec *WasmDecoder) ScBigInt {
	return ScBigInt{bytes: dec.Bytes()}
}

func BigIntEncode(enc *WasmEncoder, value ScBigInt) {
	enc.Bytes(value.bytes)
}

func BigIntFromBytes(buf []byte) ScBigInt {
	return ScBigInt{bytes: reverse(buf)}
}

func BigIntToBytes(value ScBigInt) []byte {
	return reverse(value.bytes)
}

func BigIntFromString(value string) ScBigInt {
	// Uint64 fits 18 digits or 1 quintillion
	if len(value) <= 18 {
		return NewScBigInt(Uint64FromString(value))
	}

	// build value 18 digits at a time
	digits := len(value) - 18
	lhs := BigIntFromString(value[:digits])
	rhs := BigIntFromString(value[digits:])
	return lhs.Mul(quintillion).Add(rhs)
}

func BigIntToString(value ScBigInt) string {
	if value.IsUint64() {
		return Uint64ToString(value.Uint64())
	}
	div, modulo := value.DivMod(quintillion)
	digits := Uint64ToString(modulo.Uint64())
	zeroes := "000000000000000000"[:18-len(digits)]
	return BigIntToString(div) + zeroes + digits
}

// Stupid big.Int uses BigEndian byte encoding, so our external byte encoding should
// reflect this by reverse()-ing the byte order in BigIntFromBytes and BigIntToBytes
func reverse(bytes []byte) []byte {
	n := len(bytes)
	buf := make([]byte, n)
	for i, b := range bytes {
		buf[n-1-i] = b
	}
	return buf
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
