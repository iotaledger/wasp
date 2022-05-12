package iscp

import iotago "github.com/iotaledger/iota.go/v3"

const nilAgentIDString = "-"

type NilAgentID struct {
}

var _ AgentID = &NilAgentID{}

func (a *NilAgentID) Kind() AgentIDKind {
	return AgentIDKindNil
}

func (a *NilAgentID) Bytes() []byte {
	return []byte{byte(a.Kind())}
}

func (a *NilAgentID) String(networkPrefix iotago.NetworkPrefix) string {
	return nilAgentIDString
}

func (a *NilAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	return other.Kind() == a.Kind()
}
