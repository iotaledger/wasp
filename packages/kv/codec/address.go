package codec

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

func DecodeAddress(b []byte) (ledgerstate.Address, bool, error) {
	if b == nil {
		return nil, false, nil
	}
	ret, _, err := ledgerstate.AddressFromBytes(b)
	if err != nil {
		return nil, false, err
	}
	return ret, true, nil
}

func EncodeAddress(value ledgerstate.Address) []byte {
	return value.Bytes()
}
