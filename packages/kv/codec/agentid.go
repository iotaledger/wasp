package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func DecodeAgentID(b []byte) (coretypes.AgentID, bool, error) {
	if b == nil {
		return coretypes.AgentID{}, false, nil
	}
	r, err := coretypes.NewAgentIDFromBytes(b)
	return *r, err == nil, err
}

func EncodeAgentID(value *coretypes.AgentID) []byte {
	return value.Bytes()
}
