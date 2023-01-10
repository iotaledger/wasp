package codec

import "fmt"

func DecodeString(b []byte, def ...string) (string, error) {
	if b == nil {
		if len(def) == 0 {
			return "", fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return string(b), nil
}

func MustDecodeString(b []byte, def ...string) string {
	s, err := DecodeString(b, def...)
	if err != nil {
		panic(err)
	}
	return s
}

func EncodeString(value string) []byte {
	return []byte(value)
}
