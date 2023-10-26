package codec

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/wasp/packages/hashing"
)

func DecodeHashValue(b []byte, def ...hashing.HashValue) (hashing.HashValue, error) {
	if b == nil {
		if len(def) == 0 {
			return hashing.HashValue{}, ierrors.New("cannot decode nil Hash")
		}
		return def[0], nil
	}
	return hashing.HashValueFromBytes(b)
}

func EncodeHashValue(value hashing.HashValue) []byte {
	return value[:]
}
