package codec

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
)

func DecodeColor(b []byte) (balance.Color, bool, error) {
	if b == nil {
		return balance.Color{}, false, nil
	}
	ret, _, err := balance.ColorFromBytes(b)
	return ret, err == nil, err
}

func EncodeColor(value balance.Color) []byte {
	return value[:]
}
