package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

func DecodeRequestID(b []byte) (iscp.RequestID, bool, error) {
	if b == nil {
		return iscp.RequestID{}, false, nil
	}
	r, err := iscp.RequestIDFromBytes(b)
	return r, err == nil, err
}

func EncodeRequestID(value iscp.RequestID) []byte {
	return value.Bytes()
}
