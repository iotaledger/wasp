package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func DecodeContractID(b []byte) (coretypes.ContractID, bool, error) {
	if b == nil {
		return coretypes.ContractID{}, false, nil
	}
	r, err := coretypes.NewContractIDFromBytes(b)
	return *r, err == nil, err
}

func EncodeContractID(value coretypes.ContractID) []byte {
	return value.Bytes()
}
