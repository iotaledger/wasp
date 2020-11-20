package coretypes

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"io"
)

const AgentIDLength = ChainIDLength + HnameLength

// AgentID assumes:
// - ChainID is MustAddress
// - AgentID is never used for contract with index 0  TODO ???
type AgentID [AgentIDLength]byte

func NewAgentIDFromAddress(addr address.Address) (ret AgentID) {
	copy(ret[HnameLength:], addr[:])
	return
}

func NewAgentIDFromContractID(id ContractID) (ret AgentID) {
	//if id.Hname() == 0 {
	//	panic("can't be 0 contract hname")
	//}
	copy(ret[:], id[:])
	return
}

func NewAgentIDFromBytes(data []byte) (ret AgentID, err error) {
	if len(data) != AgentIDLength {
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
	if a.IsAddress() {
		return "A-" + a.MustAddress().String()
	}
	return "C-" + a.MustContractID().String()
}

func ReadAgentID(r io.Reader, agentID *AgentID) error {
	n, err := r.Read(agentID[:])
	if err != nil {
		return err
	}
	if n != AgentIDLength {
		return errors.New("error while reading agent ID")
	}
	return nil
}
