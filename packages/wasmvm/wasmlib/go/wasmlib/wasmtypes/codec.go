// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmtypes

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// Base58Encode sandbox function wrapper for simplified use by hashtypes
var Base58Encode func(buf []byte) string

// WasmDecoder decodes separate entities from a byte buffer
type WasmDecoder struct {
	buf []byte
}

func NewWasmDecoder(buf []byte) *WasmDecoder {
	if len(buf) == 0 {
		panic("empty decode buffer")
	}
	return &WasmDecoder{buf: buf}
}

func (d *WasmDecoder) abort(msg string) {
	// make sure a deferred Close() will not trigger another panic
	d.buf = nil
	panic(msg)
}

// Byte decodes the next byte from the byte buffer
func (d *WasmDecoder) Byte() byte {
	if len(d.buf) == 0 {
		d.abort("insufficient bytes")
	}
	value := d.buf[0]
	d.buf = d.buf[1:]
	return value
}

// Bytes decodes the next variable sized slice of bytes from the byte buffer
func (d *WasmDecoder) Bytes() []byte {
	length := uint32(d.VluDecode(32))
	return d.FixedBytes(length)
}

// Close finalizes decoding by panicking if any bytes remain in the byte buffer
func (d *WasmDecoder) Close() {
	if len(d.buf) != 0 {
		d.abort("extra bytes")
	}
}

// FixedBytes decodes the next fixed size slice of bytes from the byte buffer
func (d *WasmDecoder) FixedBytes(size uint32) []byte {
	if uint32(len(d.buf)) < size {
		d.abort("insufficient fixed bytes")
	}
	value := d.buf[:size]
	d.buf = d.buf[size:]
	return value
}

// VliDecode: Variable Length Integer decoder, uses modified LEB128
func (d *WasmDecoder) VliDecode(bits int) int64 {
	b := d.Byte()
	sign := b & 0x40

	// first group of 6 bits
	value := int64(b & 0x3f)
	s := 6

	// while continuation bit is set
	for ; (b & 0x80) != 0; s += 7 {
		if s >= bits {
			d.abort("integer representation too long")
		}

		// next group of 7 bits
		b = d.Byte()
		value |= int64(b&0x7f) << s
	}

	if sign == 0 {
		// positive, sign bits are already zero
		return value
	}

	// negative, extend sign bits
	return value | (int64(-1) << s)
}

// VluDecode: Variable Length Unsigned decoder, uses ULEB128
func (d *WasmDecoder) VluDecode(bits int) uint64 {
	// first group of 7 bits
	b := d.Byte()
	value := uint64(b & 0x7f)
	s := 7

	// while continuation bit is set
	for ; (b & 0x80) != 0; s += 7 {
		if s >= bits {
			d.abort("unsigned representation too long")
		}

		// next group of 7 bits
		b = d.Byte()
		value |= uint64(b&0x7f) << s
	}
	return value
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// WasmEncoder encodes separate entities into a byte buffer
type WasmEncoder struct {
	buf []byte
}

func NewWasmEncoder() *WasmEncoder {
	return &WasmEncoder{buf: make([]byte, 0, 128)}
}

// Buf retrieves the encoded byte buffer
func (e *WasmEncoder) Buf() []byte {
	return e.buf
}

// Byte encodes a single byte into the byte buffer
func (e *WasmEncoder) Byte(value uint8) *WasmEncoder {
	e.buf = append(e.buf, value)
	return e
}

// Bytes encodes a variable sized slice of bytes into the byte buffer
func (e *WasmEncoder) Bytes(value []byte) *WasmEncoder {
	length := len(value)
	e.VluEncode(uint64(length))
	return e.FixedBytes(value, uint32(length))
}

// FixedBytes encodes a fixed size slice of bytes into the byte buffer
func (e *WasmEncoder) FixedBytes(value []byte, length uint32) *WasmEncoder {
	if uint32(len(value)) != length {
		panic("invalid fixed bytes length")
	}
	e.buf = append(e.buf, value...)
	return e
}

// VliEncode Variable Length Integer encoder, uses modified LEB128
func (e *WasmEncoder) VliEncode(value int64) *WasmEncoder {
	// bit 7 is always continuation bit

	// first group: 6 bits of data plus sign bit
	// bit 6 encodes 0 as positive and 1 as negative
	b := byte(value) & 0x3f
	value >>= 6

	finalValue := int64(0)
	if value < 0 {
		// 1st byte encodes 1 as negative in bit 6
		b |= 0x40
		// negative value, start with all high bits set to 1
		finalValue = -1
	}

	// keep shifting until all bits are done
	for value != finalValue {
		// emit with continuation bit
		e.buf = append(e.buf, b|0x80)

		// next group of 7 data bits
		b = byte(value) & 0x7f
		value >>= 7
	}

	// emit without continuation bit to signal end
	e.buf = append(e.buf, b)
	return e
}

// VluEncode Variable Length Unsigned encoder, uses ULEB128
func (e *WasmEncoder) VluEncode(value uint64) *WasmEncoder {
	// bit 7 is always continuation bit

	// first group of 7 data bits
	b := byte(value)
	value >>= 7

	// keep shifting until all bits are done
	for value != 0 {
		// emit with continuation bit
		e.buf = append(e.buf, b|0x80)

		// next group of 7 data bits
		b = byte(value)
		value >>= 7
	}

	// emit without continuation bit to signal end
	e.buf = append(e.buf, b)
	return e
}
