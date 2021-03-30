package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func DecodeRequestID(b []byte) (coretypes.RequestID, bool, error) {
	if b == nil {
		return coretypes.RequestID{}, false, nil
	}
	r, err := coretypes.RequestIDFromBytes(b)
	return r, err == nil, err
}

func EncodeRequestID(value *coretypes.RequestID) []byte {
	return value.Bytes()
}
