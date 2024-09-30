package isc

import (
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
)

const nilAgentIDString = "-"

type NilAgentID struct{}

var _ AgentID = &NilAgentID{}

func (a *NilAgentID) Bytes() []byte {
	// TODO: remove this function from codebase because it is not needed anymore
	return bcs.MustMarshal(lo.ToPtr(AgentID(a)))
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
