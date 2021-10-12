package codec

import (
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"golang.org/x/xerrors"
)

func DecodeColor(b []byte, def ...colored.Color) (colored.Color, error) {
	if b == nil {
		if len(def) == 0 {
			return colored.Color{}, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return colored.ColorFromBytes(b)
}

func EncodeColor(value colored.Color) []byte {
	return value.Bytes()
}
