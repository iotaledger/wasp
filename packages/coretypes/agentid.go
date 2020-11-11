package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

const AgentIDLenght = ChainIDLength + HnameLength

// AgentID assumes:
// - ChainID is MustAddress
// - AgentID is never used for contract with index 0  TODO ???
type AgentID [AgentIDLenght]byte

func NewAgentIDFromAddress(addr address.Address) (ret AgentID) {
	copy(ret[HnameLength:], addr[:])
	return
}

func NewAgentIDFromContractID(id ContractID) (ret AgentID) {
	if id.Hname() == 0 {
		panic("can't be 0 contract hname")
	}
	copy(ret[:], id[:])
	return
}

func NewAgentIDFromBytes(data []byte) (ret AgentID, err error) {
	if len(data) != AgentIDLenght {
		err = ErrWrongDataLength
		return
	}
	copy(ret[:], data)
	return
}

// IsAddress 0 index means it is the address
func (a AgentID) IsAddress() bool {
	return a[0] == 0 && a[1] == 0 && a[2] == 0 && a[3] == 0
}

func (a AgentID) MustAddress() (ret address.Address) {
	if !a.IsAddress() {
		panic("not an address")
	}
	copy(ret[:], a[HnameLength:])
	return
}

func (a AgentID) MustContractID() (ret ContractID) {
	if a.IsAddress() {
		panic("not a contract")
	}
	copy(ret[:], a[:])
	return
}

func (a AgentID) Bytes() []byte {
	return a[:]
}

func (a AgentID) String() string {
	panic("implement me")
}
