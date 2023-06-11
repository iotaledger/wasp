// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

type Kind byte

//////////////////// basic size-checked read/write \\\\\\\\\\\\\\\\\\\\

func ReadN(r io.Reader, data []byte) error {
	n, err := r.Read(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("incomplete read")
	}
	return nil
}

func WriteN(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("incomplete write")
	}
	return nil
}

//////////////////// bool \\\\\\\\\\\\\\\\\\\\

func ReadBool(r io.Reader) (bool, error) {
	var b [1]byte
	err := ReadN(r, b[:])
	if err != nil {
		return false, err
	}
	if (b[0] & 0xfe) != 0x00 {
		return false, errors.New("unexpected bool value")
	}
	return b[0] != 0, nil
}

func WriteBool(w io.Writer, cond bool) error {
	var b [1]byte
	if cond {
		b[0] = 1
	}
	err := WriteN(w, b[:])
	return err
}

//////////////////// byte \\\\\\\\\\\\\\\\\\\\

func ReadByte(r io.Reader) (byte, error) {
	var b [1]byte
	err := ReadN(r, b[:])
	return b[0], err
}

func WriteByte(w io.Writer, val byte) error {
	return WriteN(w, []byte{val})
}

//////////////////// bytes \\\\\\\\\\\\\\\\\\\\

func ReadBytes(r io.Reader) ([]byte, error) {
	length, err := ReadSize32(r)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return []byte{}, nil
	}
	ret := make([]byte, length)
	err = ReadN(r, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func WriteBytes(w io.Writer, data []byte) error {
	size := len(data)
	if size > math.MaxUint32 {
		panic("data size overflow")
	}
	err := WriteSize32(w, uint32(size))
	if err != nil {
		return err
	}
	if size != 0 {
		return WriteN(w, data)
	}
	return nil
}

//////////////////// int8 \\\\\\\\\\\\\\\\\\\\

func ReadInt8(r io.Reader) (int8, error) {
	val, err := ReadUint8(r)
	return int8(val), err
}

func WriteInt8(w io.Writer, val int8) error {
	return WriteUint8(w, uint8(val))
}

//////////////////// int16 \\\\\\\\\\\\\\\\\\\\

func ReadInt16(r io.Reader) (int16, error) {
	val, err := ReadUint16(r)
	return int16(val), err
}

func WriteInt16(w io.Writer, val int16) error {
	return WriteUint16(w, uint16(val))
}

//////////////////// int32 \\\\\\\\\\\\\\\\\\\\

func ReadInt32(r io.Reader) (int32, error) {
	val, err := ReadUint32(r)
	return int32(val), err
}

func WriteInt32(w io.Writer, val int32) error {
	return WriteUint32(w, uint32(val))
}

//////////////////// int64 \\\\\\\\\\\\\\\\\\\\

func ReadInt64(r io.Reader) (int64, error) {
	val, err := ReadUint64(r)
	return int64(val), err
}

func WriteInt64(w io.Writer, val int64) error {
	return WriteUint64(w, uint64(val))
}

//////////////////// size32 \\\\\\\\\\\\\\\\\\\\

func Size32FromBytes(buf []byte) (uint32, error) {
	return ReadSize32(bytes.NewReader(buf))
}

func Size32ToBytes(s uint32) []byte {
	switch {
	case s < 0x80:
		return []byte{byte(s)}
	case s < 0x4000:
		return []byte{byte(s | 0x80), byte(s >> 7)}
	case s < 0x200000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte(s >> 14)}
	case s < 0x10000000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte(s >> 21)}
	default:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte((s >> 21) | 0x80), byte(s >> 28)}
	}
}

func MustSize32FromBytes(b []byte) uint32 {
	size, err := Size32FromBytes(b)
	if err != nil {
		panic(err)
	}
	return size
}

func ReadSize32(r io.Reader) (uint32, error) {
	return decodeSize32(func() (byte, error) {
		return ReadByte(r)
	})
}

func WriteSize32(w io.Writer, value uint32) error {
	return WriteN(w, Size32ToBytes(value))
}

func decodeSize32(readByte func() (byte, error)) (uint32, error) {
	b, err := readByte()
	if err != nil {
		return 0, err
	}
	if b < 0x80 {
		return uint32(b), nil
	}
	value := uint32(b & 0x7f)

	b, err = readByte()
	if err != nil {
		return 0, err
	}
	if b < 0x80 {
		return value | (uint32(b) << 7), nil
	}
	value |= uint32(b&0x7f) << 7

	b, err = readByte()
	if err != nil {
		return 0, err
	}
	if b < 0x80 {
		return value | (uint32(b) << 14), nil
	}
	value |= uint32(b&0x7f) << 14

	b, err = readByte()
	if err != nil {
		return 0, err
	}
	if b < 0x80 {
		return value | (uint32(b) << 21), nil
	}
	value |= uint32(b&0x7f) << 21

	b, err = readByte()
	if err != nil {
		return 0, err
	}
	if b < 0xf0 {
		return value | (uint32(b) << 28), nil
	}
	return 0, errors.New("size32 overflow")
}

//////////////////// string \\\\\\\\\\\\\\\\\\\\

func ReadString(r io.Reader) (string, error) {
	ret, err := ReadBytes(r)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func WriteString(w io.Writer, str string) error {
	return WriteBytes(w, []byte(str))
}

//////////////////// uint8 \\\\\\\\\\\\\\\\\\\\

func ReadUint8(r io.Reader) (uint8, error) {
	var b [1]byte
	err := ReadN(r, b[:])
	return b[0], err
}

func WriteUint8(w io.Writer, val uint8) error {
	return WriteN(w, []byte{val})
}

//////////////////// uint16 \\\\\\\\\\\\\\\\\\\\

func ReadUint16(r io.Reader) (uint16, error) {
	var b [2]byte
	err := ReadN(r, b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b[:]), nil
}

func WriteUint16(w io.Writer, val uint16) error {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], val)
	return WriteN(w, b[:])
}

//////////////////// uint32 \\\\\\\\\\\\\\\\\\\\

func ReadUint32(r io.Reader) (uint32, error) {
	var b [4]byte
	err := ReadN(r, b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func WriteUint32(w io.Writer, val uint32) error {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], val)
	return WriteN(w, b[:])
}

//////////////////// uint64 \\\\\\\\\\\\\\\\\\\\

func ReadUint64(r io.Reader) (uint64, error) {
	var b [8]byte
	err := ReadN(r, b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

func WriteUint64(w io.Writer, val uint64) error {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], val)
	return WriteN(w, b[:])
}

//////////////////// binary marshaling \\\\\\\\\\\\\\\\\\\\

// MarshalBinary is an adapter function that uses an object's Write()
// function to marshal the object to data bytes. It is typically used
// to implement a one-line MarshalBinary() member function for the object.
func MarshalBinary(object interface{ Write(w io.Writer) error }) ([]byte, error) {
	return WriterToBytes(object), nil
}

// UnmarshalBinary is an adapter function that uses an object's Read()
// function to marshal the object from data bytes. It is typically used
// to implement a one-line UnmarshalBinary member function for the object.
func UnmarshalBinary[T interface{ Read(r io.Reader) error }](data []byte, object T) error {
	_, err := ReaderFromBytes(data, object)
	return err
}

func ReadMarshaled(r io.Reader, val encoding.BinaryUnmarshaler) error {
	if val == nil {
		panic("nil BinaryUnmarshaler")
	}
	bin, err := ReadBytes(r)
	if err != nil {
		return err
	}
	return val.UnmarshalBinary(bin)
}

func WriteMarshaled(w io.Writer, val encoding.BinaryMarshaler) error {
	if val == nil {
		panic("nil BinaryMarshaler")
	}
	bin, err := val.MarshalBinary()
	if err != nil {
		return err
	}
	return WriteBytes(w, bin)
}

//////////////////// marshalutil \\\\\\\\\\\\\\\\\\\\
//
//func FromMarshalUtil[T any](rr *Reader, fromMu func(mu *marshalutil.MarshalUtil) (T, error)) (ret T) {
//	if rr.Err == nil {
//		buf, ok := rr.r.(*bytes.Buffer)
//		if !ok {
//			panic("reader expects bytes buffer")
//		}
//		mu := marshalutil.New(buf.Bytes())
//		ret, rr.Err = fromMu(mu)
//		rr.r = bytes.NewBuffer(mu.Bytes()[mu.ReadOffset():])
//	}
//	return ret
//}

func ReaderFromMu[T interface{ Read(r io.Reader) error }](mu *marshalutil.MarshalUtil, object T) (T, error) {
	//if object == nil {
	//	panic("nil unmarshaler object")
	//}
	r := &MuReader{mu: mu}
	err := object.Read(r)
	return object, err
}

func ReadBytesFromMarshalUtil(mu *marshalutil.MarshalUtil) ([]byte, error) {
	if mu == nil {
		panic("nil MarshalUtil reader")
	}
	size, err := decodeSize32(mu.ReadByte)
	if err != nil {
		return nil, err
	}
	ret, err := mu.ReadBytes(int(size))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func WriteBytesToMarshalUtil(data []byte, mu *marshalutil.MarshalUtil) {
	if mu == nil {
		panic("nil MarshalUtil writer")
	}
	size := uint32(len(data))
	mu.WriteBytes(Size32ToBytes(size)).WriteBytes(data)
}

//////////////////// bytes \\\\\\\\\\\\\\\\\\\\

// ReadFromBytes allows a reader to use any <Type>FromBytes() function as a source.
// It will read the next group of bytes and pass it to the specified function and
// returns the correct type of object
func ReadFromBytes[T any](rr *Reader, fromBytes func([]byte) (T, error)) (ret T) {
	data := rr.ReadBytes()
	if rr.Err == nil {
		ret, rr.Err = fromBytes(data)
	}
	return ret
}

// WriteFromBytes allows a writer to use any Bytes() function as a source
func WriteFromBytes(w io.Writer, bytes interface{ Bytes() []byte }) error {
	if bytes == nil {
		panic("nil bytes writer")
	}
	return WriteN(w, bytes.Bytes())
}

// ReaderFromBytes is a wrapper that uses an object's Read() function to marshal
// the object from data bytes. It's typically used to implement a one-line
// <Type>FromBytes() function and returns the expected type and error.
func ReaderFromBytes[T interface{ Read(r io.Reader) error }](data []byte, object T) (T, error) {
	//if object == nil {
	//	panic("nil reader object")
	//}
	r := bytes.NewBuffer(data)
	if err := object.Read(r); err != nil {
		return object, err
	}
	if r.Len() != 0 {
		return object, errors.New("excess bytes")
	}
	return object, nil
}

// WriterToBytes is a wrapper that uses an object's Write() function to marshal
// the object to data bytes. It's typically used to implement a one-line Bytes()
// function for the object.
func WriterToBytes(object interface{ Write(w io.Writer) error }) []byte {
	//if object == nil {
	//	panic("nil writer object")
	//}
	w := new(bytes.Buffer)
	err := object.Write(w)
	// should never happen when writing to bytes.Buffer
	if err != nil {
		panic(err)
	}
	return w.Bytes()
}
