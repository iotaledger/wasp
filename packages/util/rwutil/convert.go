// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type Kind byte

type (
	IoReader interface {
		Read(r io.Reader) error
	}
	IoWriter interface {
		Write(w io.Writer) error
	}
	IoReadWriter interface {
		IoReader
		IoWriter
	}
)

//////////////////// basic size-checked read/write \\\\\\\\\\\\\\\\\\\\

func ReadN(r io.Reader, data []byte) error {
	n, err := io.ReadFull(r, data)
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

//////////////////// size16/size32/size64 encoding/decoding \\\\\\\\\\\\\\\\\\\\

func size16Decode(readByte func() (byte, error)) (uint16, error) {
	size64, err := size64Decode(readByte)
	if size64 >= 0x1_0000 {
		return 0, errors.New("size16 overflow")
	}
	return uint16(size64), err
}

func size32Decode(readByte func() (byte, error)) (uint32, error) {
	size64, err := size64Decode(readByte)
	if size64 >= 0x1_0000_0000 {
		return 0, errors.New("size32 overflow")
	}
	return uint32(size64), err
}

// size64Decode uses a simple variable length encoding scheme
// It takes groups of 7 bits per byte, and decodes following groups while
// the 0x80 bit is set. Since most numbers are small, this will result in
// significant storage savings, with values < 128 occupying only a single
// byte, and values < 16384 only 2 bytes.
func size64Decode(readByte func() (byte, error)) (uint64, error) {
	b, err := readByte()
	if err != nil {
		return 0, err
	}
	if b < 0x80 {
		return uint64(b), nil
	}
	value := uint64(b & 0x7f)

	for shift := 7; shift < 63; shift += 7 {
		b, err = readByte()
		if err != nil {
			return 0, err
		}
		if b < 0x80 {
			return value | (uint64(b) << shift), nil
		}
		value |= uint64(b&0x7f) << shift
	}

	// must be the final bit (since we already encoded 63 bits)
	b, err = readByte()
	if err != nil {
		return 0, err
	}
	if b > 0x01 {
		return 0, errors.New("size64 overflow")
	}
	return value | (uint64(b) << 63), nil
}

// size64Encode uses a simple variable length encoding scheme
// It takes groups of 7 bits per byte, and encodes if there will be a next group
// by setting the 0x80 bit. Since most numbers are small, this will result in
// significant storage savings, with values < 128 occupying only a single byte,
// and values < 16384 only 2 bytes.
func size64Encode(s uint64) []byte {
	// serious loop unrolling to optimize for speed
	switch {
	case s < 0x80:
		return []byte{byte(s)}
	case s < 0x4000:
		return []byte{byte(s | 0x80), byte(s >> 7)}
	case s < 0x20_0000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte(s >> 14)}
	case s < 0x1000_0000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte(s >> 21)}
	case s < 0x8_0000_0000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte((s >> 21) | 0x80), byte(s >> 28)}
	case s < 0x400_0000_0000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte((s >> 21) | 0x80), byte((s >> 28) | 0x80), byte(s >> 35)}
	case s < 0x2_0000_0000_0000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte((s >> 21) | 0x80), byte((s >> 28) | 0x80), byte((s >> 35) | 0x80), byte(s >> 42)}
	case s < 0x100_0000_0000_0000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte((s >> 21) | 0x80), byte((s >> 28) | 0x80), byte((s >> 35) | 0x80), byte((s >> 42) | 0x80), byte(s >> 49)}
	case s < 0x8000_0000_0000_0000:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte((s >> 21) | 0x80), byte((s >> 28) | 0x80), byte((s >> 35) | 0x80), byte((s >> 42) | 0x80), byte((s >> 49) | 0x80), byte(s >> 56)}
	default:
		return []byte{byte(s | 0x80), byte((s >> 7) | 0x80), byte((s >> 14) | 0x80), byte((s >> 21) | 0x80), byte((s >> 28) | 0x80), byte((s >> 35) | 0x80), byte((s >> 42) | 0x80), byte((s >> 49) | 0x80), byte((s >> 56) | 0x80), byte(s >> 63)}
	}
}

//////////////////// one-line implementation wrapper functions \\\\\\\\\\\\\\\\\\\\

// ReadFromBytes is a wrapper that uses an object's Read() function to marshal
// the object from data bytes. It's typically used to implement a one-line
// <Type>FromBytes() function and returns the expected type and error.
func ReadFromBytes[T IoReader](data []byte, obj T) (T, error) {
	// note: obj can be nil if obj.Read can handle that
	rr := NewBytesReader(data)
	rr.Read(obj)
	rr.Close()
	return obj, rr.Err
}

// WriteToBytes is a wrapper that uses an object's Write() function to marshal
// the object to data bytes. It's typically used to implement a one-line Bytes()
// function for the object.
func WriteToBytes(obj IoWriter) []byte {
	// note: obj can be nil if obj.Write can handle that
	ww := NewBytesWriter()
	ww.Write(obj)
	if ww.Err != nil {
		// should never happen when writing to Buffer
		panic(ww.Err)
	}
	return ww.Bytes()
}

//////////////////// misc generic wrapper functions \\\\\\\\\\\\\\\\\\\\

// ReadFromFunc allows a reader to use any <Type>FromBytes()-like function as a source.
// It will read the next group of bytes and pass it to the specified function and
// returns the correct type of object
func ReadFromFunc[T any](rr *Reader, fromBytes func([]byte) (T, error)) (ret T) {
	data := rr.ReadBytes()
	if rr.Err == nil {
		ret, rr.Err = fromBytes(data)
	}
	return ret
}

func BytesTest[T interface{ Bytes() []byte }](t *testing.T, obj1 T, fromBytes func(data []byte) (T, error)) T {
	obj2, err := fromBytes(obj1.Bytes())
	require.NoError(t, err)
	require.Equal(t, obj1, obj2)
	require.Equal(t, obj1.Bytes(), obj2.Bytes())
	return obj2
}

func StringTest[T interface{ String() string }](t *testing.T, obj1 T, fromString func(data string) (T, error)) T {
	obj2, err := fromString(obj1.String())
	require.NoError(t, err)
	require.Equal(t, obj1, obj2)
	require.Equal(t, obj1.String(), obj2.String())
	return obj2
}

// ReadWriteTest can be used with any object that implements IoReader and IoWriter
// to test whether the Read() and Write() functions complement each other correctly.
// You pass in an object that has all fields that need serialization set, plus a new,
// empty object that will receive the deserialized data. The function will Write, Read,
// and Write again and compare the objects for equality and both serialized versions
// as well. It will return the deserialized object in case the user wants to perform
// more tests with it.
func ReadWriteTest[T IoReadWriter](t *testing.T, obj1 T, newObj T) T {
	data1 := WriteToBytes(obj1)
	obj2, err := ReadFromBytes(data1, newObj)
	require.NoError(t, err)
	require.Equal(t, obj1, obj2)
	data2 := WriteToBytes(obj2)
	require.Equal(t, data1, data2)
	return obj2
}
