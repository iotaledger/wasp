package isc

const nilAgentIDString = "-"

type NilAgentID struct{}

var _ AgentID = &NilAgentID{}

func (a *NilAgentID) Bytes() []byte {
	return nil
}

func (a *NilAgentID) BelongsToChain(cID ChainID) bool {
	return false
}

func (a *NilAgentID) BytesWithoutChainID() []byte {
	return a.Bytes()
}

func (a *NilAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	return other.Kind() == a.Kind()
}

func (a *NilAgentID) Kind() AgentIDKind {
	return AgentIDKindNil
}

func (a *NilAgentID) String() string {
	return nilAgentIDString
}
