package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func DecodeChainID(b []byte) (coretypes.ChainID, bool, error) {
	if b == nil {
		return coretypes.ChainID{}, false, nil
	}
	ret, err := coretypes.ChainIDFromBytes(b)
	if err != nil {
		return coretypes.ChainID{}, false, err
	}
	return *ret, true, nil
}

func EncodeChainID(value coretypes.ChainID) []byte {
	return value.Bytes()
}
