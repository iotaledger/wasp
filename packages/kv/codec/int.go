package codec

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/wasp/packages/util"
)

func checkLength(d []byte, mustLen int, typeName string) error {
	if len(d) != mustLen {
		return fmt.Errorf("%d bytes expected for '%s'", mustLen, typeName)
	}
	return nil
}

func DecodeInt8(b []byte, def ...int8) (int8, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 1, "int8"); err != nil {
		return 0, err
	}
	r, err := util.Uint8From1Bytes(b)
	return int8(r), err
}

func MustDecodeInt8(b []byte, def ...int8) int8 {
	n, err := DecodeInt8(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt8(value int8) []byte {
	return util.Uint8To1Bytes(uint8(value))
}

func DecodeUint8(b []byte, def ...uint8) (uint8, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 1, "uint8"); err != nil {
		return 0, err
	}
	return util.Uint8From1Bytes(b)
}

func MustDecodeUint8(b []byte, def ...uint8) uint8 {
	n, err := DecodeUint8(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint8(value uint8) []byte {
	return util.Uint8To1Bytes(value)
}

func DecodeInt16(b []byte, def ...int16) (int16, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 2, "int16"); err != nil {
		return 0, err
	}
	r, err := util.Uint16From2Bytes(b)
	return int16(r), err
}

func MustDecodeInt16(b []byte, def ...int16) int16 {
	n, err := DecodeInt16(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt16(value int16) []byte {
	return util.Uint16To2Bytes(uint16(value))
}

func DecodeUint16(b []byte, def ...uint16) (uint16, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 2, "uint16"); err != nil {
		return 0, err
	}
	return util.Uint16From2Bytes(b)
}

func MustDecodeUint16(b []byte, def ...uint16) uint16 {
	n, err := DecodeUint16(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint16(value uint16) []byte {
	return util.Uint16To2Bytes(value)
}

func DecodeInt32(b []byte, def ...int32) (int32, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 4, "int32"); err != nil {
		return 0, err
	}
	r, err := util.Uint32From4Bytes(b)
	return int32(r), err
}

func MustDecodeInt32(b []byte, def ...int32) int32 {
	n, err := DecodeInt32(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt32(value int32) []byte {
	return util.Uint32To4Bytes(uint32(value))
}

func DecodeUint32(b []byte, def ...uint32) (uint32, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 4, "uint32"); err != nil {
		return 0, err
	}
	return util.Uint32From4Bytes(b)
}

func MustDecodeUint32(b []byte, def ...uint32) uint32 {
	n, err := DecodeUint32(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint32(value uint32) []byte {
	return util.Uint32To4Bytes(value)
}

func DecodeInt64(b []byte, def ...int64) (int64, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 8, "int64"); err != nil {
		return 0, err
	}
	r, err := util.Uint64From8Bytes(b)
	return int64(r), err
}

func MustDecodeInt64(b []byte, def ...int64) int64 {
	n, err := DecodeInt64(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeInt64(value int64) []byte {
	return util.Uint64To8Bytes(uint64(value))
}

func DecodeUint64(b []byte, def ...uint64) (uint64, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	if err := checkLength(b, 8, "uint64"); err != nil {
		return 0, err
	}
	return util.Uint64From8Bytes(b)
}

func MustDecodeUint64(b []byte, def ...uint64) uint64 {
	n, err := DecodeUint64(b, def...)
	if err != nil {
		panic(err)
	}
	return n
}

func EncodeUint64(value uint64) []byte {
	return util.Uint64To8Bytes(value)
}

func DecodeBigIntAbs(b []byte, def ...*big.Int) (*big.Int, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, fmt.Errorf("cannot decode nil bytes")
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
