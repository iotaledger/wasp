package isc

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// ContractAgentID is an AgentID formed by a ChainID and a contract Hname.
type ContractAgentID struct {
	chainID ChainID `bcs:"export"`
	hname   Hname   `bcs:"export"`
}

var _ AgentIDWithL1Address = &ContractAgentID{}

func NewContractAgentID(chainID ChainID, hname Hname) *ContractAgentID {
	return &ContractAgentID{chainID: chainID, hname: hname}
}

func contractAgentIDFromString(hnamePart, addrPart string) (AgentID, error) {
	chainID, err := ChainIDFromString(addrPart)
	if err != nil {
		return nil, fmt.Errorf("AgentIDFromString: %w", err)
	}

	h, err := HnameFromString(hnamePart)
	if err != nil {
		return nil, fmt.Errorf("AgentIDFromString: %w", err)
	}
	return NewContractAgentID(chainID, h), nil
}

func (a *ContractAgentID) Address() *cryptolib.Address {
	return a.chainID.AsAddress()
}

func (a *ContractAgentID) Bytes() []byte {
	return bcs.MustMarshal(lo.ToPtr(AgentID(a)))
}

func (a *ContractAgentID) ChainID() ChainID {
	return a.chainID
}

func (a *ContractAgentID) BelongsToChain(cID ChainID) bool {
	return a.chainID.Equals(cID)
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
	return o.chainID.Equals(a.chainID) && o.hname == a.hname
}

func (a *ContractAgentID) Hname() Hname {
	return a.hname
}

func (a *ContractAgentID) Kind() AgentIDKind {
	return AgentIDKindContract
}

func (a *ContractAgentID) String() string {
	return a.hname.String() + AgentIDStringSeparator + a.chainID.String()
}
