package codec

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/isc"
)

func DecodeAgentID(b []byte, def ...isc.AgentID) (isc.AgentID, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, xerrors.New("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return isc.AgentIDFromBytes(b)
}

func EncodeAgentID(value isc.AgentID) []byte {
	return value.Bytes()
}
