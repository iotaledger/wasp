package codec

import (
	"github.com/iotaledger/wasp/packages/isc"
	"golang.org/x/xerrors"
)

func DecodeChainID(b []byte, def ...*isc.ChainID) (*isc.ChainID, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return isc.ChainIDFromBytes(b)
}

func EncodeChainID(value *isc.ChainID) []byte {
	return value.Bytes()
}
