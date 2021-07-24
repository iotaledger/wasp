package codec

import (
	"github.com/iotaledger/wasp/packages/iscp/colored"
)

func DecodeColor(b []byte) (colored.Color, bool, error) {
	if b == nil {
		return colored.Color{}, false, nil
	}
	ret, err := colored.FromBytes(b)
	return ret, err == nil, err
}

func EncodeColor(value colored.Color) []byte {
	return value.Bytes()
}
