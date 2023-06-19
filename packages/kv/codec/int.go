package codec

import (
	"encoding/binary"
	"errors"
	"math/big"
)

func DecodeInt8(b []byte, def ...int8) (int8, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil int8")
		}
		return def[0], nil
	}
	if len(b) != 1 {
		return 0, errors.New("invalid int8 size")
	}
	return int8(b[0]), nil
}

func MustDecodeInt8(b []byte, def ...int8) int8 {
	n, err := DecodeInt8(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt8(value int8) []byte {
	return []byte{byte(value)}
}

func DecodeUint8(b []byte, def ...uint8) (uint8, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil uint8")
		}
		return def[0], nil
	}
	if len(b) != 1 {
		return 0, errors.New("invalid uint8 size")
	}
	return b[0], nil
}

func MustDecodeUint8(b []byte, def ...uint8) uint8 {
	n, err := DecodeUint8(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint8(value uint8) []byte {
	return []byte{value}
}

func DecodeInt16(b []byte, def ...int16) (int16, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil int16")
		}
		return def[0], nil
	}
	if len(b) != 2 {
		return 0, errors.New("invalid int16 size")
	}
	return int16(binary.LittleEndian.Uint16(b)), nil
}

func MustDecodeInt16(b []byte, def ...int16) int16 {
	n, err := DecodeInt16(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt16(value int16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], uint16(value))
	return b[:]
}

func DecodeUint16(b []byte, def ...uint16) (uint16, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil uint16")
		}
		return def[0], nil
	}
	if len(b) != 2 {
		return 0, errors.New("invalid uint16 size")
	}
	return binary.LittleEndian.Uint16(b), nil
}

func MustDecodeUint16(b []byte, def ...uint16) uint16 {
	n, err := DecodeUint16(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint16(value uint16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], value)
	return b[:]
}

func DecodeInt32(b []byte, def ...int32) (int32, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil int32")
		}
		return def[0], nil
	}
	if len(b) != 4 {
		return 0, errors.New("invalid int32 size")
	}
	return int32(binary.LittleEndian.Uint32(b)), nil
}

func MustDecodeInt32(b []byte, def ...int32) int32 {
	n, err := DecodeInt32(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt32(value int32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(value))
	return b[:]
}

func DecodeUint32(b []byte, def ...uint32) (uint32, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil uint32")
		}
		return def[0], nil
	}
	if len(b) != 4 {
		return 0, errors.New("invalid uint32 size")
	}
	return binary.LittleEndian.Uint32(b), nil
}

func MustDecodeUint32(b []byte, def ...uint32) uint32 {
	n, err := DecodeUint32(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint32(value uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], value)
	return b[:]
}

func DecodeInt64(b []byte, def ...int64) (int64, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil int64")
		}
		return def[0], nil
	}
	if len(b) != 8 {
		return 0, errors.New("invalid int64 size")
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

func MustDecodeInt64(b []byte, def ...int64) int64 {
	n, err := DecodeInt64(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt64(value int64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(value))
	return b[:]
}

func DecodeUint64(b []byte, def ...uint64) (uint64, error) {
	if b == nil {
		if len(def) != 1 {
			return 0, errors.New("cannot decode nil uint64")
		}
		return def[0], nil
	}
	if len(b) != 8 {
		return 0, errors.New("invalid uint64 size")
	}
	return binary.LittleEndian.Uint64(b), nil
}

func MustDecodeUint64(b []byte, def ...uint64) uint64 {
	n, err := DecodeUint64(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint64(value uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], value)
	return b[:]
}

func DecodeBigIntAbs(b []byte, def ...*big.Int) (*big.Int, error) {
	if b == nil {
		if len(def) != 1 {
			return nil, errors.New("cannot decode nil big.Int")
		}
		return def[0], nil
	}
	ret := big.NewInt(0).SetBytes(b)
	return ret, nil
}

func MustDecodeBigIntAbs(b []byte, def ...*big.Int) *big.Int {
	n, err := DecodeBigIntAbs(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeBigIntAbs(value *big.Int) []byte {
	if value == nil {
		value = big.NewInt(0)
	}
	return value.Bytes()
}
