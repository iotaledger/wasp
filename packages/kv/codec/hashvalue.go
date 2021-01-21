package codec

import (
	"github.com/iotaledger/wasp/packages/hashing"
)

// TODO return value not pointer
func DecodeHashValue(b []byte) (*hashing.HashValue, bool, error) {
	if b == nil {
		return nil, false, nil
	}
	ret, err := hashing.HashValueFromBytes(b)
	return &ret, err == nil, err
}

func EncodeHashValue(value *hashing.HashValue) []byte {
	return value[:]
}
