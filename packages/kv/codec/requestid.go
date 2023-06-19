package codec

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc"
)

func DecodeRequestID(b []byte, def ...isc.RequestID) (ret isc.RequestID, err error) {
	if b == nil {
		if len(def) == 0 {
			return ret, errors.New("cannot decode nil RequestID")
		}
		return def[0], nil
	}
	return isc.RequestIDFromBytes(b)
}

func EncodeRequestID(value isc.RequestID) []byte {
	return value.Bytes()
}
