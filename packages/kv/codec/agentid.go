package codec

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

func DecodeAgentID(b []byte) (coretypes.AgentID, bool, error) {
	if b == nil {
		return coretypes.AgentID{}, false, nil
	}
	ret, err := coretypes.NewAgentIDFromBytes(b)
	if err != nil {
		return coretypes.AgentID{}, false, err
	}
	return *ret, true, nil
}

func EncodeAgentID(value *coretypes.AgentID) []byte {
	return value.Bytes()
}
