package isc

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// EthereumAddressAgentID is an AgentID formed by an Ethereum address
type EthereumAddressAgentID struct {
	chainID ChainID        `bcs:"export"`
	eth     common.Address `bcs:"export"`
}

var _ AgentID = &EthereumAddressAgentID{}

func NewEthereumAddressAgentID(chainID ChainID, eth common.Address) *EthereumAddressAgentID {
	return &EthereumAddressAgentID{chainID: chainID, eth: eth}
}

func ethAgentIDFromString(contractPart, chainIDPart string) (*EthereumAddressAgentID, error) {
	data, err := cryptolib.DecodeHex(contractPart)
	if err != nil {
		return nil, err
	}
	if len(data) != common.AddressLength {
		return nil, errors.New("invalid ETH address string")
	}

	chainID, err := ChainIDFromString(chainIDPart)
	if err != nil {
		return nil, fmt.Errorf("invalid chainID: %w", err)
	}

	return &EthereumAddressAgentID{eth: common.BytesToAddress(data), chainID: chainID}, nil
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
	return b.eth == a.eth && b.chainID.Equals(a.chainID)
}

func (a *EthereumAddressAgentID) EthAddress() common.Address {
	return a.eth
}

func (a *EthereumAddressAgentID) ChainID() ChainID {
	return a.chainID
}

func (a *EthereumAddressAgentID) BelongsToChain(cID ChainID) bool {
	return a.chainID.Equals(cID)
}

func (a *EthereumAddressAgentID) BytesWithoutChainID() []byte {
	return a.eth[:]
}

func (a *EthereumAddressAgentID) Kind() AgentIDKind {
	return AgentIDKindEthereumAddress
}

func (a *EthereumAddressAgentID) String() string {
	// eth.String includes 0x prefix
	return a.eth.String() + AgentIDStringSeparator + a.chainID.String()
}
