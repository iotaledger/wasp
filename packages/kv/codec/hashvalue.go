package codec

import (
	"github.com/iotaledger/wasp/packages/hashing"
)

func DecodeHashValue(b []byte) (hashing.HashValue, bool, error) {
	if b == nil {
		return hashing.NilHash, false, nil
	}
	r, err := hashing.HashValueFromBytes(b)
	return r, err == nil, err
}

func EncodeHashValue(value hashing.HashValue) []byte {
	return value[:]
}
