package codec

import (
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

func DecodeInt8(b []byte, def ...int8) (int8, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	r, err := util.Uint8From1Bytes(b)
	return int8(r), err
}

func EncodeInt8(value int8) []byte {
	return util.Uint8To1Bytes(uint8(value))
}

func DecodeUint8(b []byte, def ...uint8) (uint8, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return util.Uint8From1Bytes(b)
}

func EncodeUint8(value uint8) []byte {
	return util.Uint8To1Bytes(value)
}

func DecodeInt16(b []byte, def ...int16) (int16, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	r, err := util.Uint16From2Bytes(b)
	return int16(r), err
}

func EncodeInt16(value int16) []byte {
	return util.Uint16To2Bytes(uint16(value))
}

func DecodeUint16(b []byte, def ...uint16) (uint16, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return util.Uint16From2Bytes(b)
}

func EncodeUint16(value uint16) []byte {
	return util.Uint16To2Bytes(value)
}

func DecodeInt32(b []byte, def ...int32) (int32, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	r, err := util.Uint32From4Bytes(b)
	return int32(r), err
}

func EncodeInt32(value int32) []byte {
	return util.Uint32To4Bytes(uint32(value))
}

func DecodeUint32(b []byte, def ...uint32) (uint32, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return util.Uint32From4Bytes(b)
}

func EncodeUint32(value uint32) []byte {
	return util.Uint32To4Bytes(value)
}

func DecodeInt64(b []byte, def ...int64) (int64, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	r, err := util.Uint64From8Bytes(b)
	return int64(r), err
}

func EncodeInt64(value int64) []byte {
	return util.Uint64To8Bytes(uint64(value))
}

func DecodeUint64(b []byte, def ...uint64) (uint64, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return util.Uint64From8Bytes(b)
}

func EncodeUint64(value uint64) []byte {
	return util.Uint64To8Bytes(value)
}
