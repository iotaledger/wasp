package codec

import "errors"

var Bool = NewCodec(decodeBool, encodeBool)

func decodeBool(b []byte) (bool, error) {
	if len(b) != 1 {
		return false, errors.New("invalid bool size")
	}
	if (b[0] & 0xfe) != 0x00 {
		return false, errors.New("invalid bool value")
	}
	return b[0] != 0, nil
}

func encodeBool(value bool) []byte {
	if value {
		return []byte{1}
	}
	return []byte{0}
}
