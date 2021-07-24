package codec

import (
	"github.com/iotaledger/wasp/packages/iscp/color"
)

func DecodeColor(b []byte) (color.Color, bool, error) {
	if b == nil {
		return color.Color{}, false, nil
	}
	ret, err := color.FromBytes(b)
	return ret, err == nil, err
}

func EncodeColor(value color.Color) []byte {
	return value.Bytes()
}
