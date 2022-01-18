// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type BytesDecoder struct {
	data []byte
}

func NewBytesDecoder(data []byte) *BytesDecoder {
	if len(data) == 0 {
		panic("cannot decode empty byte array, use exist()")
	}
	return &BytesDecoder{data: data}
}

func (d *BytesDecoder) Address() ScAddress {
	return NewScAddressFromBytes(d.FixedBytes(uint32(TypeSizes[TYPE_ADDRESS])))
}

func (d *BytesDecoder) AgentID() ScAgentID {
	return NewScAgentIDFromBytes(d.FixedBytes(uint32(TypeSizes[TYPE_AGENT_ID])))
}

func (d *BytesDecoder) Bool() bool {
	return d.Uint8() != 0
}

func (d *BytesDecoder) Bytes() []byte {
	length := d.Uint32()
	return d.FixedBytes(length)
}

func (d *BytesDecoder) ChainID() ScChainID {
	return NewScChainIDFromBytes(d.FixedBytes(uint32(TypeSizes[TYPE_CHAIN_ID])))
}

func (d *BytesDecoder) Close() {
	if len(d.data) != 0 {
		panic("extra bytes")
	}
}

func (d *BytesDecoder) Color() ScColor {
	return NewScColorFromBytes(d.FixedBytes(uint32(TypeSizes[TYPE_COLOR])))
}

func (d *BytesDecoder) FixedBytes(size uint32) []byte {
	if uint32(len(d.data)) < size {
		panic("insufficient bytes")
	}
	value := d.data[:size]
	d.data = d.data[size:]
	return value
}

func (d *BytesDecoder) Hash() ScHash {
	return NewScHashFromBytes(d.FixedBytes(uint32(TypeSizes[TYPE_HASH])))
}

func (d *BytesDecoder) Hname() ScHname {
	return NewScHnameFromBytes(d.FixedBytes(uint32(TypeSizes[TYPE_HNAME])))
}

func (d *BytesDecoder) Int8() int8 {
	return int8(d.Uint8())
}

func (d *BytesDecoder) Int16() int16 {
	return int16(d.vliDecode(16))
}

func (d *BytesDecoder) Int32() int32 {
	return int32(d.vliDecode(32))
}

func (d *BytesDecoder) Int64() int64 {
	return d.vliDecode(64)
}

func (d *BytesDecoder) RequestID() ScRequestID {
	return NewScRequestIDFromBytes(d.FixedBytes(uint32(TypeSizes[TYPE_REQUEST_ID])))
}

func (d *BytesDecoder) String() string {
	return string(d.Bytes())
}

func (d *BytesDecoder) Uint8() uint8 {
	if len(d.data) == 0 {
		panic("insufficient bytes")
	}
	value := d.data[0]
	d.data = d.data[1:]
	return value
}

func (d *BytesDecoder) Uint16() uint16 {
	return uint16(d.vluDecode(16))
}

func (d *BytesDecoder) Uint32() uint32 {
	return uint32(d.vluDecode(32))
}

func (d *BytesDecoder) Uint64() uint64 {
	return d.vluDecode(64)
}

// vli (variable length integer) decoder
func (d *BytesDecoder) vliDecode(bits int) (value int64) {
	b := d.Uint8()
	sign := b & 0x40

	// first group of 6 bits
	value = int64(b & 0x3f)
	s := 6

	// while continuation bit is set
	for (b & 0x80) != 0 {
		if s >= bits {
			panic("integer representation too long")
		}

		// next group of 7 bits
		b = d.Uint8()
		value |= int64(b&0x7f) << s
		s += 7
	}

	if sign == 0 {
		// positive, sign bits are already zero
		return value
	}

	// negative, extend sign bits
	return value | (int64(-1) << s)
}

// vlu (variable length unsigned) decoder
func (d *BytesDecoder) vluDecode(bits int) uint64 {
	// first group of 7 bits
	b := d.Uint8()
	value := uint64(b & 0x7f)
	s := 7

	// while continuation bit is set
	for (b & 0x80) != 0 {
		if s >= bits {
			panic("integer representation too long")
		}

		// next group of 7 bits
		b = d.Uint8()
		value |= uint64(b&0x7f) << s
		s += 7
	}
	return value
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type BytesEncoder struct {
	data []byte
}

func NewBytesEncoder() *BytesEncoder {
	return &BytesEncoder{data: make([]byte, 0, 128)}
}

func (e *BytesEncoder) Address(value ScAddress) *BytesEncoder {
	return e.FixedBytes(value.Bytes(), uint32(TypeSizes[TYPE_ADDRESS]))
}

func (e *BytesEncoder) AgentID(value ScAgentID) *BytesEncoder {
	return e.FixedBytes(value.Bytes(), uint32(TypeSizes[TYPE_AGENT_ID]))
}

func (e *BytesEncoder) Bool(value bool) *BytesEncoder {
	if value {
		return e.Uint8(1)
	}
	return e.Uint8(0)
}

func (e *BytesEncoder) Bytes(value []byte) *BytesEncoder {
	length := uint32(len(value))
	e.Uint32(length)
	e.FixedBytes(value, length)
	return e
}

func (e *BytesEncoder) ChainID(value ScChainID) *BytesEncoder {
	return e.FixedBytes(value.Bytes(), uint32(TypeSizes[TYPE_CHAIN_ID]))
}

func (e *BytesEncoder) Color(value ScColor) *BytesEncoder {
	return e.FixedBytes(value.Bytes(), uint32(TypeSizes[TYPE_COLOR]))
}

func (e *BytesEncoder) Data() []byte {
	return e.data
}

func (e *BytesEncoder) FixedBytes(value []byte, length uint32) *BytesEncoder {
	if len(value) != int(length) {
		panic("invalid fixed bytes length")
	}
	e.data = append(e.data, value...)
	return e
}

func (e *BytesEncoder) Hash(value ScHash) *BytesEncoder {
	return e.FixedBytes(value.Bytes(), uint32(TypeSizes[TYPE_HASH]))
}

func (e *BytesEncoder) Hname(value ScHname) *BytesEncoder {
	return e.FixedBytes(value.Bytes(), uint32(TypeSizes[TYPE_HNAME]))
}

func (e *BytesEncoder) Int8(value int8) *BytesEncoder {
	return e.Uint8(uint8(value))
}

func (e *BytesEncoder) Int16(value int16) *BytesEncoder {
	return e.Int64(int64(value))
}

func (e *BytesEncoder) Int32(value int32) *BytesEncoder {
	return e.Int64(int64(value))
}

// vli (variable length integer) encoder
func (e *BytesEncoder) Int64(value int64) *BytesEncoder {
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
		e.data = append(e.data, b|0x80)

		// next group of 7 bits
		b = byte(value) & 0x7f
		value >>= 7
	}

	// emit without continuation bit
	e.data = append(e.data, b)
	return e
}

func (e *BytesEncoder) RequestID(value ScRequestID) *BytesEncoder {
	return e.FixedBytes(value.Bytes(), uint32(TypeSizes[TYPE_REQUEST_ID]))
}

func (e *BytesEncoder) String(value string) *BytesEncoder {
	return e.Bytes([]byte(value))
}

func (e *BytesEncoder) Uint8(value uint8) *BytesEncoder {
	e.data = append(e.data, value)
	return e
}

func (e *BytesEncoder) Uint16(value uint16) *BytesEncoder {
	return e.Uint64(uint64(value))
}

func (e *BytesEncoder) Uint32(value uint32) *BytesEncoder {
	return e.Uint64(uint64(value))
}

// vlu (variable length unsigned) encoder
func (e *BytesEncoder) Uint64(value uint64) *BytesEncoder {
	// first group of 7 bits
	b := byte(value)
	value >>= 7

	// keep shifting until all bits are done
	for value != 0 {
		// emit with continuation bit
		e.data = append(e.data, b|0x80)

		// next group of 7 bits
		b = byte(value)
		value >>= 7
	}

	// emit without continuation bit
	e.data = append(e.data, b)
	return e
}
