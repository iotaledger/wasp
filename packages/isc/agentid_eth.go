package isc

import (
	"bytes"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// EthereumAddressAgentID is an AgentID formed by an Ethereum address
type EthereumAddressAgentID struct {
	eth common.Address `bcs:"export"`
}

var _ AgentID = &EthereumAddressAgentID{}

func NewEthereumAddressAgentID(eth common.Address) *EthereumAddressAgentID {
	return &EthereumAddressAgentID{eth: eth}
}

func ethAgentIDFromString(contractPart string) (*EthereumAddressAgentID, error) {
	data, err := cryptolib.DecodeHex(contractPart)
	if err != nil {
		return nil, err
	}
	if len(data) != common.AddressLength {
		return nil, errors.New("invalid ETH address string")
	}

	return &EthereumAddressAgentID{eth: common.BytesToAddress(data)}, nil
}

func (a *EthereumAddressAgentID) Bytes() []byte {
	return bcs.MustMarshal(lo.ToPtr(AgentID(a)))
}

func (a *EthereumAddressAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	if other.Kind() != a.Kind() {
		return false
	}
	b := other.(*EthereumAddressAgentID)
	return bytes.Equal(b.eth.Bytes(), a.eth.Bytes())
}

func (a *EthereumAddressAgentID) EthAddress() common.Address {
	return a.eth
}

func (a *EthereumAddressAgentID) Kind() AgentIDKind {
	return AgentIDKindEthereumAddress
}

func (a *EthereumAddressAgentID) String() string {
	// eth.String includes 0x prefix
	return a.eth.String()
}
