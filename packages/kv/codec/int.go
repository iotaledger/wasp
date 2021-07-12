package codec

import (
	"github.com/iotaledger/wasp/packages/util"
)

func DecodeInt16(b []byte) (int16, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return 0, false, nil
	}

	r, err := util.Uint16From2Bytes(b)
	return int16(r), err == nil, err
}

func EncodeInt16(value int16) []byte {
	return util.Uint16To2Bytes(uint16(value))
}

func DecodeUint16(b []byte) (uint16, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return 0, false, nil
	}

	r, err := util.Uint16From2Bytes(b)
	return r, err == nil, err
}

func EncodeUint16(value uint16) []byte {
	return util.Uint16To2Bytes(value)
}

func DecodeInt32(b []byte) (int32, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return 0, false, nil
	}

	r, err := util.Uint32From4Bytes(b)
	return int32(r), err == nil, err
}

func EncodeInt32(value int32) []byte {
	return util.Uint32To4Bytes(uint32(value))
}

func DecodeUint32(b []byte) (uint32, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return 0, false, nil
	}

	r, err := util.Uint32From4Bytes(b)
	return r, err == nil, err
}

func EncodeUint32(value uint32) []byte {
	return util.Uint32To4Bytes(value)
}

func DecodeInt64(b []byte) (int64, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return 0, false, nil
	}

	r, err := util.Uint64From8Bytes(b)
	return int64(r), err == nil, err
}

func EncodeInt64(value int64) []byte {
	return util.Uint64To8Bytes(uint64(value))
}

func DecodeUint64(b []byte) (uint64, bool, error) {
	if b == nil {
		// special behavior for backward compatibility: nil value is treated as the absence of a value, not an error
		return 0, false, nil
	}

	r, err := util.Uint64From8Bytes(b)
	return r, err == nil, err
}

func EncodeUint64(value uint64) []byte {
	return util.Uint64To8Bytes(value)
}
