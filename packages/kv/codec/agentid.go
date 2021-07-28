package codec

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

func DecodeAgentID(b []byte) (iscp.AgentID, bool, error) {
	if b == nil {
		return iscp.AgentID{}, false, nil
	}
	ret, err := iscp.AgentIDFromBytes(b)
	if err != nil {
		return iscp.AgentID{}, false, err
	}
	return *ret, true, nil
}

func EncodeAgentID(value *iscp.AgentID) []byte {
	return value.Bytes()
}
