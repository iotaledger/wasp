package codec

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/isc"
)

func DecodeChainID(b []byte, def ...isc.ChainID) (isc.ChainID, error) {
	if b == nil {
		if len(def) == 0 {
			return isc.ChainID{}, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return isc.ChainIDFromBytes(b)
}

func EncodeChainID(chainID isc.ChainID) []byte {
	return chainID.Bytes()
}
