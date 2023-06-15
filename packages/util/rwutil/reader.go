// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"bytes"
	"encoding"
	"errors"
	"io"
	"math"
	"math/big"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
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
	return NewReader(bytes.NewBuffer(data))
}

// PushBack returns a pushback writer that allows you to insert data before the stream.
// The Reader will read this data first, and then resume reading from the stream.
// The pushback Writer is only valid for this Reader until it resumes the stream.
func (rr *Reader) PushBack() *Writer {
	push := &PushBack{rr: rr, r: rr.r, buf: new(bytes.Buffer)}
	rr.r = push
	return &Writer{w: push}
}

func (rr *Reader) Read(reader interface{ Read(r io.Reader) error }) {
	if reader == nil {
		panic("nil reader")
	}
	if rr.Err == nil {
		rr.Err = reader.Read(rr.r)
	}
}

func (rr *Reader) ReadN(ret []byte) {
	if rr.Err == nil {
		rr.Err = ReadN(rr.r, ret)
	}
}

func (rr *Reader) ReadBool() (ret bool) {
	if rr.Err == nil {
		ret, rr.Err = ReadBool(rr.r)
	}
	return ret
}

//nolint:govet
func (rr *Reader) ReadByte() (ret byte) {
	if rr.Err == nil {
		ret, rr.Err = ReadByte(rr.r)
	}
	return ret
}

func (rr *Reader) ReadBytes() (ret []byte) {
	if rr.Err == nil {
		ret, rr.Err = ReadBytes(rr.r)
	}
	return ret
}

func (rr *Reader) ReadDuration() (ret time.Duration) {
	return time.Duration(rr.ReadInt64())
}

func (rr *Reader) ReadInt8() (ret int8) {
	if rr.Err == nil {
		ret, rr.Err = ReadInt8(rr.r)
	}
	return ret
}

func (rr *Reader) ReadInt16() (ret int16) {
	if rr.Err == nil {
		ret, rr.Err = ReadInt16(rr.r)
	}
	return ret
}

func (rr *Reader) ReadInt32() (ret int32) {
	if rr.Err == nil {
		ret, rr.Err = ReadInt32(rr.r)
	}
	return ret
}

func (rr *Reader) ReadInt64() (ret int64) {
	if rr.Err == nil {
		ret, rr.Err = ReadInt64(rr.r)
	}
	return ret
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

func (rr *Reader) ReadMarshaled(m encoding.BinaryUnmarshaler) {
	if m == nil {
		panic("nil unmarshaler")
	}
	buf := rr.ReadBytes()
	if rr.Err == nil {
		rr.Err = m.UnmarshalBinary(buf)
	}
}

type deserializable interface {
	Deserialize([]byte, serializer.DeSerializationMode, interface{}) (int, error)
}

// ReadSerialized reads the deserializable object from the stream.
// If no sizes are present a 16-bit size is read from the stream.
// The first size indicates a different limit for the size read from the stream.
// The second size indicates the expected size and does not read it from the stream.
func (rr *Reader) ReadSerialized(s deserializable, sizes ...int) {
	if rr.Err != nil {
		return
	}
	if s == nil {
		panic("nil deserializer")
	}
	var size int
	switch len(sizes) {
	case 0:
		size = rr.ReadSize16()
	case 1:
		limit := sizes[0]
		if limit < 0 || limit > math.MaxInt32 {
			panic("invalid deserialize limit")
		}
		size = rr.ReadSizeWithLimit(uint32(limit))
	case 2:
		size = sizes[1]
		if size < 0 || size > math.MaxInt32 {
			panic("invalid deserialize size")
		}
	default:
		panic("too many deserialize params")
	}
	data := make([]byte, size)
	rr.ReadN(data)
	if rr.Err == nil {
		var n int
		n, rr.Err = s.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
		if n != len(data) && rr.Err == nil {
			rr.Err = errors.New("unexpected deserialize size")
		}
	}
}

func (rr *Reader) ReadSize16() (ret int) {
	return rr.ReadSizeWithLimit(math.MaxUint16)
}

func (rr *Reader) ReadSize32() (ret int) {
	// note we cannot exceed SIGNED max
	// because if int is actually 32 bit it would turn negative
	return rr.ReadSizeWithLimit(math.MaxInt32)
}

func (rr *Reader) ReadSizeWithLimit(limit uint32) int {
	if rr.Err != nil {
		return 0
	}
	var size32 uint32
	size32, rr.Err = ReadSize32(rr.r)
	if size32 > limit && rr.Err == nil {
		rr.Err = errors.New("read size limit overflow")
	}
	return int(size32)
}

func (rr *Reader) ReadString() (ret string) {
	if rr.Err == nil {
		ret, rr.Err = ReadString(rr.r)
	}
	return ret
}

func (rr *Reader) ReadUint8() (ret uint8) {
	if rr.Err == nil {
		ret, rr.Err = ReadUint8(rr.r)
	}
	return ret
}

func (rr *Reader) ReadUint16() (ret uint16) {
	if rr.Err == nil {
		ret, rr.Err = ReadUint16(rr.r)
	}
	return ret
}

func (rr *Reader) ReadUint32() (ret uint32) {
	if rr.Err == nil {
		ret, rr.Err = ReadUint32(rr.r)
	}
	return ret
}

func (rr *Reader) ReadUint64() (ret uint64) {
	if rr.Err == nil {
		ret, rr.Err = ReadUint64(rr.r)
	}
	return ret
}

func (rr *Reader) ReadUint256() (ret *big.Int) {
	ret = new(big.Int)
	ret.SetBytes(rr.ReadBytes())
	return ret
}
