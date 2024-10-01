package isc

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
)

// AddressAgentID is an AgentID backed by a L1 address
type AddressAgentID struct {
	a *cryptolib.Address `bcs:""`
}

var _ AgentIDWithL1Address = &AddressAgentID{}

func NewAddressAgentID(addr *cryptolib.Address) *AddressAgentID {
	return &AddressAgentID{a: addr}
}

func addressAgentIDFromString(s string) (*AddressAgentID, error) {
	addr, err := cryptolib.NewAddressFromHexString(s)
	if err != nil {
		return nil, err
	}
	return &AddressAgentID{a: addr}, nil
}

func (a *AddressAgentID) Address() *cryptolib.Address {
	return a.a
}

func (a *AddressAgentID) Bytes() []byte {
	// TODO: remove this function from codebase because it is not needed anymore
	return bcs.MustMarshal(lo.ToPtr(AgentID(a)))
}

func (a *AddressAgentID) BelongsToChain(ChainID) bool {
	return false
}

func (a *AddressAgentID) BytesWithoutChainID() []byte {
	return a.Bytes()
}

func (a *AddressAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	if other.Kind() != a.Kind() {
		return false
	}
	return other.(*AddressAgentID).a.Equals(a.a)
}

func (a *AddressAgentID) Kind() AgentIDKind {
	return AgentIDKindAddress
}

func (a *AddressAgentID) String() string {
	return a.a.String()
}
