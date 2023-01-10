package codec

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
)

func DecodeHashValue(b []byte, def ...hashing.HashValue) (hashing.HashValue, error) {
	if b == nil {
		if len(def) == 0 {
			return hashing.HashValue{}, fmt.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return hashing.HashValueFromBytes(b)
}

func EncodeHashValue(value hashing.HashValue) []byte {
	return value[:]
}
