package codec

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

// Deprecated: use DecodeColor instead
func DecodeColorLedgerstate(b []byte) (ledgerstate.Color, bool, error) {
	if b == nil {
		return ledgerstate.Color{}, false, nil
	}
	ret, _, err := ledgerstate.ColorFromBytes(b)
	return ret, err == nil, err
}

// Deprecated: use EncodeColor instead
func EncodeColorLedgerstate(value ledgerstate.Color) []byte {
	return value.Bytes()
}
