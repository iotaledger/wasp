package codec

import (
	"errors"
)

func DecodeBool(b []byte, def ...bool) (bool, error) {
	if b == nil {
		if len(def) == 0 {
			return false, errors.New("cannot decode nil bool")
		}
		return def[0], nil
	}
	if len(b) != 1 {
		return false, errors.New("invalid bool size")
	}
	if (b[0] & 0xfe) != 0x00 {
		return false, errors.New("invalid bool value")
	}
	return b[0] != 0, nil
}

func MustDecodeBool(b []byte, def ...bool) bool {
	ret, err := DecodeBool(b, def...)
	if err != nil {
		panic(err)
	}
	return ret
}

func EncodeBool(value bool) []byte {
	if value {
		return []byte{1}
	}
	return []byte{0}
}
