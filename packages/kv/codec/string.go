package codec

import "golang.org/x/xerrors"

func DecodeString(b []byte, def ...string) (string, error) {
	if b == nil {
		if len(def) == 0 {
			return "", xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return string(b), nil
}

func EncodeString(value string) []byte {
	return []byte(value)
}
