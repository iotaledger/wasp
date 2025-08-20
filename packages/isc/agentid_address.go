package isc

import (
	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// AddressAgentID is an AgentID backed by a L1 address
type AddressAgentID struct {
	a *cryptolib.Address `bcs:"export"`
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
	return bcs.MustMarshal(lo.ToPtr(AgentID(a)))
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
