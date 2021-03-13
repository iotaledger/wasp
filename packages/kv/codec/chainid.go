package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func DecodeChainID(b []byte) (coretypes.ChainID, bool, error) {
	if b == nil {
		return coretypes.ChainID{}, false, nil
	}
	r, err := coretypes.NewChainIDFromBytes(b)
	return r, err == nil, err
}

func EncodeChainID(value coretypes.ChainID) []byte {
	return value.Bytes()
}
