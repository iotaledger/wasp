// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"bytes"
	"encoding"
	"errors"
	"io"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

type Reader struct {
	Err error
	r   io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

func NewBytesReader(data []byte) *Reader {
	return NewReader(bytes.NewBuffer(data))
}

// PushBack returns a pushback writer that allows you to insert data before the stream.
// The Reader will read this data first, and then resume reading from the stream.
// The pushback Writer is only valid for this Reader until it resumes the stream.
func (rr *Reader) PushBack() *Writer {
	pb := &PushBack{rr: rr, r: rr.r, buf: new(bytes.Buffer)}
	rr.r = pb
	return NewWriter(pb.buf)
}

func (rr *Reader) Read(reader interface{ Read(r io.Reader) error }) {
	if rr.Err == nil {
		rr.Err = reader.Read(rr.r)
	}
}

func (rr *Reader) ReadN(ret []byte) {
	if rr.Err == nil {
		rr.Err = ReadN(rr.r, ret)
	}
}

func (rr *Reader) ReadAddress() (ret iotago.Address) {
	addrType := rr.ReadByte()
	if rr.Err != nil {
		return ret
	}
	ret, rr.Err = iotago.AddressSelector(uint32(addrType))
	if rr.Err != nil {
		return ret
	}
	buf := make([]byte, ret.Size())
	buf[0] = addrType
	rr.ReadN(buf[1:])
	if rr.Err != nil {
		return ret
	}
	_, rr.Err = ret.Deserialize(buf, serializer.DeSeriModeNoValidation, nil)
	return ret
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

func (rr *Reader) ReadMarshaled(m encoding.BinaryUnmarshaler) {
	buf := rr.ReadBytes()
	if rr.Err == nil {
		if m == nil {
			rr.Err = errors.New("nil unmarshaler")
			return
		}
		rr.Err = m.UnmarshalBinary(buf)
	}
}

type deserializable interface {
	Deserialize([]byte, serializer.DeSerializationMode, interface{}) (int, error)
}

func (rr *Reader) ReadSerialized(s deserializable) {
	data := rr.ReadBytes()
	if rr.Err == nil {
		if s == nil {
			rr.Err = errors.New("nil deserializer")
			return
		}
		var n int
		n, rr.Err = s.Deserialize(data, serializer.DeSeriModeNoValidation, nil)
		if rr.Err == nil && n != len(data) {
			rr.Err = errors.New("incomplete deserialize")
		}
	}
}

func (rr *Reader) ReadSize() (ret int) {
	return int(rr.ReadSize32())
}

func (rr *Reader) ReadSize32() (ret uint32) {
	if rr.Err == nil {
		ret, rr.Err = ReadSize32(rr.r)
	}
	return ret
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
