package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

func DecodeRequestID(b []byte, def ...iscp.RequestID) (iscp.RequestID, error) {
	if b == nil {
		if len(def) == 0 {
			return iscp.RequestID{}, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return iscp.RequestIDFromBytes(b)
}

func EncodeRequestID(value iscp.RequestID) []byte {
	return value.Bytes()
}
