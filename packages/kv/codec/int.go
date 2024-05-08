package codec

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"golang.org/x/exp/constraints"
)

var (
	Int8      = newInt8Codec[int8]()
	Uint8     = newInt8Codec[uint8]()
	Int16     = newIntCodec[int16](2, binary.LittleEndian.Uint16, binary.LittleEndian.PutUint16)
	Uint16    = newIntCodec[uint16](2, binary.LittleEndian.Uint16, binary.LittleEndian.PutUint16)
	Int32     = newIntCodec[int32](4, binary.LittleEndian.Uint32, binary.LittleEndian.PutUint32)
	Uint32    = newIntCodec[uint32](4, binary.LittleEndian.Uint32, binary.LittleEndian.PutUint32)
	Int64     = newIntCodec[int64](8, binary.LittleEndian.Uint64, binary.LittleEndian.PutUint64)
	Uint64    = newIntCodec[uint64](8, binary.LittleEndian.Uint64, binary.LittleEndian.PutUint64)
	BigIntAbs = NewCodec(decodeBigIntAbs, encodeBigIntAbs)
)

func newInt8Codec[T constraints.Integer]() Codec[T] {
	return NewCodec(
		func(b []byte) (r T, err error) {
			if len(b) != 1 {
				return 0, fmt.Errorf("%T: bytes length must be 1", r)
			}
			return T(b[0]), nil
		},
		func(value T) []byte {
			return []byte{byte(value)}
		},
	)
}

func newIntCodec[T constraints.Integer, U constraints.Unsigned](size int, dec func([]byte) U, enc func([]byte, U)) Codec[T] {
	return NewCodec(
		func(b []byte) (r T, err error) {
			if len(b) != size {
				return 0, fmt.Errorf("%T: bytes length must be %d", r, size)
			}
			return T(dec(b)), nil
		},
		func(value T) []byte {
			b := make([]byte, size)
			enc(b, U(value))
			return b[:]
		},
	)
}

func decodeBigIntAbs(b []byte) (*big.Int, error) {
	ret := big.NewInt(0).SetBytes(b)
	return ret, nil
}

func encodeBigIntAbs(value *big.Int) []byte {
	if value == nil {
		value = big.NewInt(0)
	}
	return value.Bytes()
}
