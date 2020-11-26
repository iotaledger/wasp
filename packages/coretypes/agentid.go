package coretypes

import (
	"errors"
	"io"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/mr-tron/base58"
)

const AgentIDLength = ChainIDLength + HnameLength

// AgentID assumes:
// - ChainID is MustAddress
// - AgentID is never used for contract with index 0
type AgentID [AgentIDLength]byte

// NewAgentIDFromAddress a constructor
func NewAgentIDFromAddress(addr address.Address) (ret AgentID) {
	copy(ret[HnameLength:], addr[:])
	return
}

// NewAgentIDFromContractID a constructor
func NewAgentIDFromContractID(id ContractID) (ret AgentID) {
	copy(ret[:], id[:])
	return
}

// NewAgentIDFromBytes a constructor
func NewAgentIDFromBytes(data []byte) (ret AgentID, err error) {
	if len(data) != AgentIDLength {
		err = ErrWrongDataLength
		return
	}
	copy(ret[:], data)
	return
}

// IsAddress checks if agentID represents address. 0 index means it is the address
func (a AgentID) IsAddress() bool {
	return a[0] == 0 && a[1] == 0 && a[2] == 0 && a[3] == 0
}

// MustAddress takes address or panic if not address
func (a AgentID) MustAddress() (ret address.Address) {
	if !a.IsAddress() {
		panic("not an address")
	}
	copy(ret[:], a[HnameLength:])
	return
}

// MustContractID takes contract ID or panics if not a contract ID
func (a AgentID) MustContractID() (ret ContractID) {
	if a.IsAddress() {
		panic("not a contract")
	}
	copy(ret[:], a[:])
	return
}

// Bytes marshals to bytes
func (a AgentID) Bytes() []byte {
	return a[:]
}

// String human readable string
func (a AgentID) String() string {
	if a.IsAddress() {
		return "A-" + a.MustAddress().String()
	}
	return "C-" + a.MustContractID().String()
}

// ReadAgentID reading/unmarshaling
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

func (a AgentID) Base58() string {
	return base58.Encode(a[:])
}

func AgentIDFromBase58(s string) (ret AgentID, err error) {
	var data []byte
	if data, err = base58.Decode(s); err != nil {
		return
	}
	return NewAgentIDFromBytes(data)
}
