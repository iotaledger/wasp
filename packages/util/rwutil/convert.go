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

//////////////////// size32 encoding/decoding \\\\\\\\\\\\\\\\\\\\

func size32Decode(readByte func() (byte, error)) (uint32, error) {
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

func size32Encode(s uint32) []byte {
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

//////////////////// one-line implementation wrapper functions \\\\\\\\\\\\\\\\\\\\

// ReadFromBytes is a wrapper that uses an object's Read() function to marshal
// the object from data bytes. It's typically used to implement a one-line
// <Type>FromBytes() function and returns the expected type and error.
func ReadFromBytes[T IoReader](data []byte, obj T) (T, error) {
	// note: obj can be nil if obj.Read can handle that
	rr := NewBytesReader(data)
	rr.Read(obj)
	if rr.Err != nil {
		return obj, rr.Err
	}
	if len(rr.Bytes()) != 0 {
		return obj, errors.New("excess bytes in buffer")
	}
	return obj, nil
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

// SerializationTest can be used with any object that implements IoReader and IoWriter
// to test whether the Read() and Write() functions complement each other correctly.
// You pass in an object that has all fields that need serialization set, plus a new,
// empty object that will receive the deserialized data. The function will Write, Read,
// and Write again and compare the objects for equality and both serialized versions
// as well.
func SerializationTest[T IoReadWriter](t *testing.T, obj1 T, newObj T) {
	data1 := WriteToBytes(obj1)
	obj2, err := ReadFromBytes(data1, newObj)
	require.NoError(t, err)
	require.Equal(t, obj1, obj2)
	data2 := WriteToBytes(obj2)
	require.Equal(t, data1, data2)
}
