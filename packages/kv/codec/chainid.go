package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

func DecodeChainID(b []byte, def ...*iscp.ChainID) (*iscp.ChainID, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}

	chainId, err := iscp.ChainIDFromBytes(b)

	return chainId, err
}

func EncodeChainID(value *iscp.ChainID) []byte {
	return value.Bytes()
}
