// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

type BytesDecoder struct {
	data []byte
}

func NewBytesDecoder(data []byte) *BytesDecoder {
	return &BytesDecoder{data: data}
}

func (d *BytesDecoder) Bool() bool {
	return d.Int8() != 0
}

func (d *BytesDecoder) Bytes() []byte {
	size := int(d.Int32())
	if len(d.data) < size {
		panic("insufficient bytes")
	}
	value := d.data[:size]
	d.data = d.data[size:]
	return value
}

func (d *BytesDecoder) Int8() int8 {
	if len(d.data) == 0 {
		panic("insufficient bytes")
	}
	value := d.data[0]
	d.data = d.data[1:]
	return int8(value)
}

func (d *BytesDecoder) Int16() int16 {
	return int16(d.leb128Decode(16))
}

func (d *BytesDecoder) Int32() int32 {
	return int32(d.leb128Decode(32))
}

func (d *BytesDecoder) Int64() int64 {
	return d.leb128Decode(64)
}

// leb128 decoder
func (d *BytesDecoder) leb128Decode(bits int) int64 {
	val := int64(0)
	s := 0
	for {
		if len(d.data) == 0 {
			panic("insufficient bytes")
		}
		b := int8(d.data[0])
		d.data = d.data[1:]
		val |= int64(b&0x7f) << s
		if (b & -0x80) == 0 {
			if int8(val>>s)&0x7f != b&0x7f {
				panic("integer too large")
			}

			// extend int7 sign to int8
			b |= (b & 0x40) << 1

			// extend int8 sign to int64
			return val | (int64(b) << s)
		}
		s += 7
		if s >= bits {
			panic("integer representation too long")
		}
	}
}

func (d *BytesDecoder) Uint8() uint8 {
	return uint8(d.Int8())
}

func (d *BytesDecoder) Uint16() uint16 {
	return uint16(d.Int16())
}

func (d *BytesDecoder) Uint32() uint32 {
	return uint32(d.Int32())
}

func (d *BytesDecoder) Uint64() uint64 {
	return uint64(d.Int64())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type BytesEncoder struct {
	data []byte
}

func NewBytesEncoder() *BytesEncoder {
	return &BytesEncoder{data: make([]byte, 0, 128)}
}

func (e *BytesEncoder) Bool(value bool) *BytesEncoder {
	if value {
		return e.Int8(1)
	}
	return e.Int8(0)
}

func (e *BytesEncoder) Bytes(value []byte) *BytesEncoder {
	e.Int32(int32(len(value)))
	e.data = append(e.data, value...)
	return e
}

func (e *BytesEncoder) Data() []byte {
	return e.data
}

func (e *BytesEncoder) Int8(value int8) *BytesEncoder {
	e.data = append(e.data, byte(value))
	return e
}

func (e *BytesEncoder) Int16(value int16) *BytesEncoder {
	return e.leb128Encode(int64(value))
}

func (e *BytesEncoder) Int32(value int32) *BytesEncoder {
	return e.leb128Encode(int64(value))
}

func (e *BytesEncoder) Int64(value int64) *BytesEncoder {
	return e.leb128Encode(value)
}

// leb128 encoder
func (e *BytesEncoder) leb128Encode(value int64) *BytesEncoder {
	for {
		b := byte(value)
		s := b & 0x40
		value >>= 7
		if (value == 0 && s == 0) || (value == -1 && s != 0) {
			e.data = append(e.data, b&0x7f)
			return e
		}
		e.data = append(e.data, b|0x80)
	}
}

func (e *BytesEncoder) Uint8(value uint8) *BytesEncoder {
	return e.Int8(int8(value))
}

func (e *BytesEncoder) Uint16(value uint16) *BytesEncoder {
	return e.Int16(int16(value))
}

func (e *BytesEncoder) Uint32(value uint32) *BytesEncoder {
	return e.Int32(int32(value))
}

func (e *BytesEncoder) Uint64(value uint64) *BytesEncoder {
	return e.Int64(int64(value))
}
