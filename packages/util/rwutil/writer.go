// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"runtime/debug"
	"time"

	"fortio.org/safecast"
)

type Writer struct {
	Err error
	w   io.Writer
}

func NewWriter(w io.Writer) *Writer {
	if w == nil {
		panic("nil io.Writer")
	}
	return &Writer{w: w}
}

func NewBytesWriter() *Writer {
	// We're about to write one or more items.
	// Pre-allocate a reasonable-sized buffer to prevent excessive copying.
	// After that fills up, Go's append() will take care of growing.
	buf := make(Buffer, 0, 128)
	return NewWriter(&buf)
}

// Bytes will return the accumulated bytes in the Writer buffer.
// It is asserted that the write stream is a Buffer, and that
// the Writer error state is nil.
func (ww *Writer) Bytes() []byte {
	buf, ok := ww.w.(*Buffer)
	if !ok {
		panic("writer expects bytes buffer")
	}
	if ww.Err != nil {
		// WTF? writing to Buffer never fails
		panic(ww.Err)
	}
	return *buf
}

func (ww *Writer) Skip() *Reader {
	skip := &Skipper{ww: ww, w: ww.w}
	ww.w = skip
	return &Reader{r: skip}
}

func (ww *Writer) Write(obj IoWriter) *Writer {
	if ww.Err == nil {
		ww.Err = obj.Write(ww.w)
	}
	return ww
}

func (ww *Writer) WriteN(val []byte) *Writer {
	if ww.Err == nil {
		ww.Err = WriteN(ww.w, val)
	}
	return ww
}

func (ww *Writer) WriteUint256(val *big.Int) *Writer {
	if ww.Err == nil {
		if val == nil {
			val = new(big.Int)
		}
		if val.Sign() >= 0 {
			return ww.WriteBytes(val.Bytes())
		}
		ww.Err = errors.New("negative uint256")
	}
	return ww
}

// WriteAmount16 writes a variable-length encoded amount.
func (ww *Writer) WriteAmount16(val uint16) *Writer {
	if ww.Err == nil {
		ww.WriteN(size64Encode(uint64(val)))
	}
	return ww
}

// WriteAmount32 writes a variable-length encoded amount.
func (ww *Writer) WriteAmount32(val uint32) *Writer {
	if ww.Err == nil {
		ww.WriteN(size64Encode(uint64(val)))
	}
	return ww
}

// WriteAmount64 writes a variable-length encoded amount.
func (ww *Writer) WriteAmount64(val uint64) *Writer {
	if ww.Err == nil {
		ww.WriteN(size64Encode(val))
	}
	return ww
}

func (ww *Writer) WriteBool(val bool) *Writer {
	if ww.Err == nil {
		data := []byte{0x00}
		if val {
			data[0] = 0x01
		}
		ww.Err = WriteN(ww.w, data)
	}
	return ww
}

//nolint:govet
func (ww *Writer) WriteByte(val byte) *Writer {
	if ww.Err == nil {
		ww.Err = WriteN(ww.w, []byte{val})
	}
	return ww
}

func (ww *Writer) WriteBytes(data []byte) *Writer {
	if ww.Err == nil {
		size := len(data)
		if size > math.MaxInt32 {
			panic("data size overflow")
		}
		ww.WriteSize32(size)
		if size != 0 {
			ww.WriteN(data)
		}
	}
	return ww
}

func (ww *Writer) WriteDuration(val time.Duration) *Writer {
	uint64Val, err := safecast.Convert[uint64](val)
	if err != nil {
		ww.Err = err
		return ww
	}

	return ww.WriteUint64(uint64Val)
}

func (ww *Writer) WriteFromBytes(obj interface{ Bytes() []byte }) *Writer {
	if ww.Err == nil {
		ww.WriteBytes(obj.Bytes())
	}
	return ww
}

func (ww *Writer) WriteFromFunc(write func(w io.Writer) (int, error)) *Writer {
	if ww.Err == nil {
		_, ww.Err = write(ww.w)
	}
	return ww
}

// WriteGas64 writes a variable-length encoded amount of gas.
// Note that the amount is incremented before storing so that the
// math.MaxUint64 gas limit will wrap to zero and only takes 1 byte.
func (ww *Writer) WriteGas64(val uint64) *Writer {
	return ww.WriteAmount64(val + 1)
}

func (ww *Writer) WriteInt8(val int8) *Writer {
	uint8Val, err := safecast.Convert[uint8](val)
	if err != nil {
		ww.Err = err
		return ww
	}

	return ww.WriteUint8(uint8Val)
}

func (ww *Writer) WriteInt16(val int16) *Writer {
	uint16Val, err := safecast.Convert[uint16](val)
	if err != nil {
		ww.Err = err
		return ww
	}

	return ww.WriteUint16(uint16Val)
}

func (ww *Writer) WriteInt32(val int32) *Writer {
	uint32Val, err := safecast.Convert[uint32](val)
	if err != nil {
		ww.Err = err
		return ww
	}

	return ww.WriteUint32(uint32Val)
}

func (ww *Writer) WriteInt64(val int64) *Writer {
	uint64Val, err := safecast.Convert[uint64](val)
	if err != nil {
		ww.Err = err
		return ww
	}

	return ww.WriteUint64(uint64Val)
}

func (ww *Writer) WriteKind(msgType Kind) *Writer {
	return ww.WriteByte(byte(msgType))
}

func (ww *Writer) WriteSize16(val int) *Writer {
	return ww.WriteSizeWithLimit(val, math.MaxUint16)
}

func (ww *Writer) WriteSize32(val int) *Writer {
	// note we cannot exceed SIGNED max
	// because if int is actually 32 bit it would become negative
	return ww.WriteSizeWithLimit(val, math.MaxInt32)
}

func (ww *Writer) WriteSizeWithLimit(val int, limit uint32) *Writer {
	if ww.Err == nil {
		if 0 <= val && val <= int(limit) {
			return ww.WriteN(size64Encode(uint64(val)))
		}
		debug.PrintStack()
		ww.Err = fmt.Errorf("invalid write size limit (val: %d, limit: %d)", val, limit)
	}
	return ww
}

func (ww *Writer) WriteString(val string) *Writer {
	return ww.WriteBytes([]byte(val))
}

func (ww *Writer) WriteUint8(val uint8) *Writer {
	if ww.Err == nil {
		ww.Err = WriteN(ww.w, []byte{val})
	}
	return ww
}

func (ww *Writer) WriteUint16(val uint16) *Writer {
	if ww.Err == nil {
		ww.Err = WriteN(ww.w, []byte{byte(val), byte(val >> 8)})
	}
	return ww
}

func (ww *Writer) WriteUint32(val uint32) *Writer {
	if ww.Err == nil {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], val)
		ww.Err = WriteN(ww.w, b[:])
	}
	return ww
}

func (ww *Writer) WriteUint64(val uint64) *Writer {
	if ww.Err == nil {
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], val)
		ww.Err = WriteN(ww.w, b[:])
	}
	return ww
}

func (ww *Writer) WriteBigUint(val *big.Int) *Writer {
	if ww.Err == nil {
		if val == nil {
			val = new(big.Int)
		}
		if val.Sign() >= 0 {
			return ww.WriteBytes(val.Bytes())
		}
		ww.Err = errors.New("negative BigInt")
	}
	return ww
}
