// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package rwutil provides utilities for reading and writing operations.
package rwutil

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/big"
	"time"
)

type Reader struct {
	Err error
	r   io.Reader
}

func NewReader(r io.Reader) *Reader {
	if r == nil {
		panic("nil io.Reader")
	}
	return &Reader{r: r}
}

func NewBytesReader(data []byte) *Reader {
	buf := Buffer(data)
	return &Reader{r: &buf}
}

// Bytes will return the remaining bytes in the Reader buffer.
// It is asserted that the read stream is a Buffer, and that
// the Reader error state is nil.
func (rr *Reader) Bytes() []byte {
	buf, ok := rr.r.(*Buffer)
	if !ok {
		panic("reader expects bytes buffer")
	}
	if rr.Err != nil {
		panic(rr.Err)
	}
	return *buf
}

func (rr *Reader) CheckAvailable(nrOfBytes int) int {
	if rr.Err != nil {
		return 0
	}
	if buf, ok := rr.r.(*Buffer); ok && len(*buf) < nrOfBytes {
		rr.Err = errors.New("insufficient bytes remaining in buffer")
		return 0
	}
	return nrOfBytes
}

func (rr *Reader) Available() int {
	buff, ok := rr.r.(*Buffer)
	if !ok {
		panic("reader is not a buffer")
	}

	return buff.Size()
}

// Close indicates the end of reading from the bytes buffer.
// If any unread bytes are remaining in the buffer an error will be returned.
func (rr *Reader) Close() {
	if rr.Err == nil && len(rr.Bytes()) != 0 {
		rr.Err = errors.New("excess bytes in buffer")
	}
}

// Must will wrap the reader stream in a stream that will panic whenever an error occurs.
func (rr *Reader) Must() *Reader {
	must := &Must{r: rr.r}
	rr.r = must
	return rr
}

// PushBack returns a pushback writer that allows you to insert data before the stream.
// The Reader will read this data first, and then resume reading from the stream.
// The pushback Writer is only valid for this Reader until it resumes the stream.
func (rr *Reader) PushBack() *Writer {
	push := &PushBack{rr: rr, r: rr.r}
	rr.r = push
	return &Writer{w: push}
}

func (rr *Reader) Read(obj IoReader) {
	if rr.Err == nil {
		rr.Err = obj.Read(rr.r)
	}
}

func (rr *Reader) ReadN(ret []byte) {
	if rr.Err == nil {
		rr.Err = ReadN(rr.r, ret)
	}
}

// ReadAmount16 reads a variable-length encoded amount.
func (rr *Reader) ReadAmount16() (ret uint16) {
	if rr.Err == nil {
		ret, rr.Err = size16Decode(func() (byte, error) {
			return rr.ReadByte(), rr.Err
		})
	}
	return ret
}

// ReadAmount32 reads a variable-length encoded amount.
func (rr *Reader) ReadAmount32() (ret uint32) {
	if rr.Err == nil {
		ret, rr.Err = size32Decode(func() (byte, error) {
			return rr.ReadByte(), rr.Err
		})
	}
	return ret
}

// ReadAmount64 reads a variable-length encoded amount.
func (rr *Reader) ReadAmount64() (ret uint64) {
	if rr.Err == nil {
		ret, rr.Err = size64Decode(func() (byte, error) {
			return rr.ReadByte(), rr.Err
		})
	}
	return ret
}

func (rr *Reader) ReadBool() bool {
	if rr.Err != nil {
		return false
	}
	var b [1]byte
	rr.Err = ReadN(rr.r, b[:])
	if (b[0] & 0xfe) != 0x00 {
		rr.Err = errors.New("unexpected bool value")
		return false
	}
	return b[0] != 0x00
}

//nolint:govet
func (rr *Reader) ReadByte() byte {
	if rr.Err != nil {
		return 0
	}
	var b [1]byte
	rr.Err = ReadN(rr.r, b[:])
	return b[0]
}

func (rr *Reader) ReadBytes() []byte {
	if rr.Err != nil {
		return nil
	}
	size := rr.ReadSize32()
	if rr.Err != nil {
		return nil
	}
	if size == 0 {
		return []byte{}
	}
	ret := make([]byte, size)
	rr.Err = ReadN(rr.r, ret)
	if rr.Err != nil {
		return nil
	}
	return ret
}

func (rr *Reader) ReadDuration() (ret time.Duration) {
	return time.Duration(rr.ReadUint64())
}

func (rr *Reader) ReadFromFunc(read func(w io.Reader) (int, error)) {
	if rr.Err == nil {
		_, rr.Err = read(rr.r)
	}
}

// ReadGas64 reads a variable-length encoded amount of gas.
// Note that the amount was incremented before storing so that the
// math.MaxUint64 gas limit will wrap to zero and only takes 1 byte.
func (rr *Reader) ReadGas64() (ret uint64) {
	ret = rr.ReadAmount64()
	if rr.Err == nil {
		ret--
	}
	return ret
}

func (rr *Reader) ReadInt8() (ret int8) {
	return int8(rr.ReadUint8())
}

func (rr *Reader) ReadInt16() (ret int16) {
	return int16(rr.ReadUint16())
}

func (rr *Reader) ReadInt32() (ret int32) {
	return int32(rr.ReadUint32())
}

func (rr *Reader) ReadInt64() (ret int64) {
	return int64(rr.ReadUint64())
}

func (rr *Reader) ReadKind() Kind {
	return Kind(rr.ReadByte())
}

func (rr *Reader) ReadKindAndVerify(expectedKind Kind) {
	kind := rr.ReadKind()
	if kind != expectedKind && rr.Err == nil {
		rr.Err = errors.New("unexpected object kind")
	}
}

// ReadSize16 reads a 16-bit size from the stream.
// We expect this size to indicate how many items we are about to read
// from the stream. Therefore, if we can determine that there are not
// at least this amount of bytes available in the stream we raise an
// error and return zero for the size.
func (rr *Reader) ReadSize16() (size int) {
	size = rr.ReadSizeWithLimit(math.MaxUint16)
	return rr.CheckAvailable(size)
}

// ReadSize32 reads a 32-bit size from the stream.
// We expect this size to indicate how many items we are about to read
// from the stream. Therefore, if we can determine that there are not
// at least this amount of bytes available in the stream we raise an
// error and return zero for the size.
func (rr *Reader) ReadSize32() (size int) {
	// Note that we cannot exceed SIGNED math.MaxInt32, because we don't
	// want the returned int to turn negative in case ints are 32 bits
	size = rr.ReadSizeWithLimit(math.MaxInt32)
	return rr.CheckAvailable(size)
}

// ReadSizeWithLimit reads an int size from the stream, and verifies that
// it does not exceed the specified limit. By limiting the size we can
// better detect malformed input data. The size returned will always be
// zero if an error occurred.
func (rr *Reader) ReadSizeWithLimit(limit uint32) int {
	if rr.Err != nil {
		return 0
	}
	var size32 uint32
	size32, rr.Err = size32Decode(func() (byte, error) {
		return rr.ReadByte(), rr.Err
	})
	if size32 > limit && rr.Err == nil {
		rr.Err = errors.New("read size limit overflow")
		return 0
	}
	return int(size32)
}

func (rr *Reader) ReadString() (ret string) {
	return string(rr.ReadBytes())
}

func (rr *Reader) ReadUint8() (ret uint8) {
	return rr.ReadByte()
}

func (rr *Reader) ReadUint16() (ret uint16) {
	if rr.Err != nil {
		return 0
	}
	var b [2]byte
	rr.Err = ReadN(rr.r, b[:])
	if rr.Err != nil {
		return 0
	}
	return uint16(b[0]) | (uint16(b[1]) << 8)
}

func (rr *Reader) ReadUint32() (ret uint32) {
	if rr.Err != nil {
		return 0
	}
	var b [4]byte
	rr.Err = ReadN(rr.r, b[:])
	if rr.Err != nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b[:])
}

func (rr *Reader) ReadUint64() (ret uint64) {
	if rr.Err != nil {
		return 0
	}
	var b [8]byte
	rr.Err = ReadN(rr.r, b[:])
	if rr.Err != nil {
		return 0
	}
	return binary.LittleEndian.Uint64(b[:])
}

func (rr *Reader) ReadBigUint() (ret *big.Int) {
	ret = new(big.Int)
	data := rr.ReadBytes()
	if data != nil {
		ret.SetBytes(data)
	}
	return ret
}

func ReadStruct[T IoReader](rr *Reader, v T) T {
	rr.Read(v)
	return v
}
