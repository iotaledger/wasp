package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
)

func DecodeChainID(b []byte) (chainid.ChainID, bool, error) {
	if b == nil {
		return chainid.ChainID{}, false, nil
	}
	ret, err := chainid.ChainIDFromBytes(b)
	if err != nil {
		return chainid.ChainID{}, false, err
	}
	return *ret, true, nil
}

func EncodeChainID(value chainid.ChainID) []byte {
	return value.Bytes()
}
