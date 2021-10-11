package codec

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"golang.org/x/xerrors"
)

func DecodeAddress(b []byte, def ...ledgerstate.Address) (ledgerstate.Address, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	ret, _, err := ledgerstate.AddressFromBytes(b)
	return ret, err
}

func EncodeAddress(value ledgerstate.Address) []byte {
	return value.Bytes()
}
