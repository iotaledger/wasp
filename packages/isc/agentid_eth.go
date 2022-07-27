package isc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/hive.go/marshalutil"
)

// EthereumAddressAgentID is an AgentID formed by an Ethereum address
type EthereumAddressAgentID struct {
	eth common.Address
}

var _ AgentID = &EthereumAddressAgentID{}

func NewEthereumAddressAgentID(eth common.Address) *EthereumAddressAgentID {
	return &EthereumAddressAgentID{eth: eth}
}

func ethAgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (AgentID, error) {
	var ethBytes []byte
	var err error
	if ethBytes, err = mu.ReadBytes(common.AddressLength); err != nil {
		return nil, err
	}
	var eth common.Address
	eth.SetBytes(ethBytes)
	return NewEthereumAddressAgentID(eth), nil
}

func ethAgentIDFromString(s string) (AgentID, error) {
	eth := common.HexToAddress(s)
	return NewEthereumAddressAgentID(eth), nil
}

func (a *EthereumAddressAgentID) EthAddress() common.Address {
	return a.eth
}

func (a *EthereumAddressAgentID) Kind() AgentIDKind {
	return AgentIDKindEthereumAddress
}

func (a *EthereumAddressAgentID) Bytes() []byte {
	return append([]byte{byte(a.Kind())}, a.eth.Bytes()...)
}

func (a *EthereumAddressAgentID) String() string {
	return a.eth.String() // includes "0x"
}

func (a *EthereumAddressAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	if other.Kind() != a.Kind() {
		return false
	}
	e := other.(*EthereumAddressAgentID)
	return e.eth == a.eth
}
