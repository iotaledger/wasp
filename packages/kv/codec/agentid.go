package codec

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc"
)

var ErrNilAgentID = errors.New("cannot decode nil AgentID")

func DecodeAgentID(b []byte, def ...isc.AgentID) (isc.AgentID, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, ErrNilAgentID
		}
		return def[0], nil
	}
	return isc.AgentIDFromBytes(b)
}

func MustDecodeAgentID(b []byte, def ...isc.AgentID) isc.AgentID {
	r, err := DecodeAgentID(b, def...)
	if err != nil {
		panic(err)
	}
	return r
}

func EncodeAgentID(value isc.AgentID) []byte {
	return value.Bytes()
}
