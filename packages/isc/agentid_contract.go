package isc

import (
	"fmt"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// ContractAgentID is an AgentID formed by a contract Hname.
type ContractAgentID struct {
	hname Hname `bcs:"export"`
}

var _ AgentIDWithL1Address = &ContractAgentID{}

func NewContractAgentID(hname Hname) *ContractAgentID {
	return &ContractAgentID{hname: hname}
}

func contractAgentIDFromString(hnamePart string) (AgentID, error) {
	h, err := HnameFromString(hnamePart)
	if err != nil {
		return nil, fmt.Errorf("AgentIDFromString: %w", err)
	}
	return NewContractAgentID(h), nil
}

func (a *ContractAgentID) Address() *cryptolib.Address {
	return cryptolib.NewEmptyAddress()
}

func (a *ContractAgentID) Bytes() []byte {
	return bcs.MustMarshal(lo.ToPtr(AgentID(a)))
}

func (a *ContractAgentID) BytesWithoutChainID() []byte {
	return bcs.MustMarshal(&a.hname)
}

func (a *ContractAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	if other.Kind() != a.Kind() {
		return false
	}
	o := other.(*ContractAgentID)
	return o.hname == a.hname
}

func (a *ContractAgentID) Hname() Hname {
	return a.hname
}

func (a *ContractAgentID) Kind() AgentIDKind {
	return AgentIDKindContract
}

func (a *ContractAgentID) String() string {
	k := a.hname.String()
	return k
}
