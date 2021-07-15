package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

func DecodeChainID(b []byte) (iscp.ChainID, bool, error) {
	if b == nil {
		return iscp.ChainID{}, false, nil
	}
	ret, err := iscp.ChainIDFromBytes(b)
	if err != nil {
		return iscp.ChainID{}, false, err
	}
	return *ret, true, nil
}

func EncodeChainID(value iscp.ChainID) []byte {
	return value.Bytes()
}
