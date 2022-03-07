package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

func DecodeAgentID(b []byte, def ...*iscp.AgentID) (*iscp.AgentID, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, xerrors.New("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return iscp.AgentIDFromBytes(b)
}

func EncodeAgentID(value *iscp.AgentID) []byte {
	return value.Bytes()
}
