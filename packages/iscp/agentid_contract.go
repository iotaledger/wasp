package iscp

import (
	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"golang.org/x/xerrors"
)

// ContractAgentID is an AgentID formed by a ChainID and a contract Hname.
type ContractAgentID struct {
	chainID *ChainID
	h       Hname
}

var _ AgentIDWithL1Address = &ContractAgentID{}

func NewContractAgentID(chainID *ChainID, hname Hname) *ContractAgentID {
	return &ContractAgentID{chainID: chainID, h: hname}
}

func contractAgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (AgentID, error) {
	var chID *ChainID
	var err error
	if chID, err = ChainIDFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	var h Hname
	if h, err = HnameFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return NewContractAgentID(chID, h), nil
}

func contractAgentIDFromString(hnamePart, addrPart string) (AgentID, error) {
	chID, err := ChainIDFromString(addrPart)
	if err != nil {
		return nil, xerrors.Errorf("NewAgentIDFromString: %v", err)
	}
	h, err := HnameFromString(hnamePart)
	if err != nil {
		return nil, xerrors.Errorf("NewAgentIDFromString: %v", err)
	}
	return NewContractAgentID(chID, h), nil
}

func (a *ContractAgentID) Address() iotago.Address {
	return a.chainID.AsAddress()
}

func (a *ContractAgentID) ChainID() *ChainID {
	return a.chainID
}

func (a *ContractAgentID) Hname() Hname {
	return a.h
}

func (a *ContractAgentID) Kind() AgentIDKind {
	return AgentIDKindContract
}

func (a *ContractAgentID) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(byte(a.Kind()))
	mu.WriteBytes(a.chainID.Bytes())
	mu.WriteBytes(a.h.Bytes())
	return mu.Bytes()
}

func (a *ContractAgentID) String() string {
	return a.h.String() + "@" + a.chainID.String()
}

func (a *ContractAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	if other.Kind() != a.Kind() {
		return false
	}
	o := other.(*ContractAgentID)
	return o.chainID.Equals(a.chainID) && o.h == a.h
}
