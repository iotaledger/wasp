// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmcodec

type WasmDecoder struct {
	buf []byte
}

func NewWasmDecoder(buf []byte) *WasmDecoder {
	return &WasmDecoder{buf: buf}
}

func (d *WasmDecoder) abort(msg string) {
	// make sure deferred Close() will not trigger another panic
	d.buf = nil
	panic(msg)
}

func (d *WasmDecoder) Byte() byte {
	if len(d.buf) == 0 {
		d.abort("insufficient bytes")
	}
	value := d.buf[0]
	d.buf = d.buf[1:]
	return value
}

func (d *WasmDecoder) Bytes() []byte {
	length := uint32(d.VluDecode(32))
	return d.FixedBytes(length)
}

func (d *WasmDecoder) Close() {
	if len(d.buf) != 0 {
		d.abort("extra bytes")
	}
}

func (d *WasmDecoder) FixedBytes(size uint32) []byte {
	if uint32(len(d.buf)) < size {
		d.abort("insufficient bytes")
	}
	value := d.buf[:size]
	d.buf = d.buf[size:]
	return value
}

// VliDecode: Variable Length Integer decoder
func (d *WasmDecoder) VliDecode(bits int) (value int64) {
	b := d.Byte()
	sign := b & 0x40

	// first group of 6 bits
	value = int64(b & 0x3f)
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

// VluDecode: Variable Length Unsigned decoder
func (d *WasmDecoder) VluDecode(bits int) uint64 {
	// first group of 7 bits
	b := d.Byte()
	value := uint64(b & 0x7f)
	s := 7

	// while continuation bit is set
	for ; (b & 0x80) != 0; s += 7 {
		if s >= bits {
			d.abort("integer representation too long")
		}

		// next group of 7 bits
		b = d.Byte()
		value |= uint64(b&0x7f) << s
	}
	return value
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type WasmEncoder struct {
	buf []byte
}

func NewWasmEncoder() *WasmEncoder {
	return &WasmEncoder{buf: make([]byte, 0, 128)}
}

func (e *WasmEncoder) Buf() []byte {
	return e.buf
}

func (e *WasmEncoder) Byte(value uint8) *WasmEncoder {
	e.buf = append(e.buf, value)
	return e
}

func (e *WasmEncoder) Bytes(value []byte) *WasmEncoder {
	length := len(value)
	e.VluEncode(uint64(length))
	e.FixedBytes(value, uint32(length))
	return e
}

func (e *WasmEncoder) FixedBytes(value []byte, length uint32) *WasmEncoder {
	if len(value) != int(length) {
		panic("invalid fixed bytes length")
	}
	e.buf = append(e.buf, value...)
	return e
}

// vli (variable length integer) encoder
func (e *WasmEncoder) VliEncode(value int64) *WasmEncoder {
	// first group of 6 bits
	// 1st byte encodes 0 as positive in bit 6
	b := byte(value) & 0x3f
	value >>= 6

	finalValue := int64(0)
	if value < 0 {
		// encode negative value
		// 1st byte encodes 1 as negative in bit 6
		b |= 0x40
		finalValue = -1
	}

	// keep shifting until all bits are done
	for value != finalValue {
		// emit with continuation bit
		e.buf = append(e.buf, b|0x80)

		// next group of 7 bits
		b = byte(value) & 0x7f
		value >>= 7
	}

	// emit without continuation bit
	e.buf = append(e.buf, b)
	return e
}

// vlu (variable length unsigned) encoder
func (e *WasmEncoder) VluEncode(value uint64) *WasmEncoder {
	// first group of 7 bits
	b := byte(value)
	value >>= 7

	// keep shifting until all bits are done
	for value != 0 {
		// emit with continuation bit
		e.buf = append(e.buf, b|0x80)

		// next group of 7 bits
		b = byte(value)
		value >>= 7
	}

	// emit without continuation bit
	e.buf = append(e.buf, b)
	return e
}
