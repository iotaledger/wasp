package coret

import (
	"bytes"
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

// NewAgentIDFromContractID a constructor
func NewAgentIDFromContractID(id ContractID) (ret AgentID) {
	copy(ret[:], id[:])
	return
}

// NewAgentIDFromAddress a constructor
func NewAgentIDFromAddress(addr address.Address) AgentID {
	// 0 is a reserved hname
	return NewAgentIDFromContractID(NewContractID(ChainID(addr), 0))
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

func NewRandomAgentID() AgentID {
	chainID := NewRandomChainID()
	hname := Hn("testFunction")
	return NewAgentIDFromContractID(NewContractID(chainID, hname))
}

func (a *AgentID) chainIDField() []byte {
	return a[:ChainIDLength]
}

func (a *AgentID) hnameField() []byte {
	return a[ChainIDLength : ChainIDLength+HnameLength]
}

// IsAddress checks if agentID represents address. 0 in the place of the contract's hname means it is an address
func (a AgentID) IsAddress() bool {
	var z [4]byte
	return bytes.Equal(a.hnameField(), z[:])
}

// MustAddress takes address or panic if not address
func (a AgentID) MustAddress() (ret address.Address) {
	if !a.IsAddress() {
		panic("not an address")
	}
	copy(ret[:], a.chainIDField())
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
