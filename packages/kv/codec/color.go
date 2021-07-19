package codec

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

func DecodeColor(b []byte) (ledgerstate.Color, bool, error) {
	if b == nil {
		return ledgerstate.Color{}, false, nil
	}
	ret, _, err := ledgerstate.ColorFromBytes(b)
	return ret, err == nil, err
}

func EncodeColor(value ledgerstate.Color) []byte {
	return value.Bytes()
}
