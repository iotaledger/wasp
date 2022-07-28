package isc

const nilAgentIDString = "-"

type NilAgentID struct{}

var _ AgentID = &NilAgentID{}

func (a *NilAgentID) Kind() AgentIDKind {
	return AgentIDKindNil
}

func (a *NilAgentID) Bytes() []byte {
	return []byte{byte(a.Kind())}
}

func (a *NilAgentID) String() string {
	return nilAgentIDString
}

func (a *NilAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	return other.Kind() == a.Kind()
}
