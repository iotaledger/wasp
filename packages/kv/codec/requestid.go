package codec

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc"
)

func DecodeRequestID(b []byte, def ...isc.RequestID) (isc.RequestID, error) {
	if b == nil {
		if len(def) == 0 {
			return isc.RequestID{}, errors.New("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return isc.RequestIDFromBytes(b)
}

func EncodeRequestID(value isc.RequestID) []byte {
	return value.Bytes()
}
