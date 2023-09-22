package isc

import (
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// EthereumAddressAgentID is an AgentID formed by an Ethereum address
type EthereumAddressAgentID struct {
	chainID ChainID
	eth     common.Address
}

var _ AgentID = &EthereumAddressAgentID{}

func NewEthereumAddressAgentID(chainID ChainID, eth common.Address) *EthereumAddressAgentID {
	return &EthereumAddressAgentID{chainID: chainID, eth: eth}
}

func ethAgentIDFromString(contractPart, chainIDPart string) (*EthereumAddressAgentID, error) {
	data, err := iotago.DecodeHex(contractPart)
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
	return rwutil.WriteToBytes(a)
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

func (a *EthereumAddressAgentID) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(rwutil.Kind(a.Kind()))
	rr.Read(&a.chainID)
	rr.ReadN(a.eth[:])
	return rr.Err
}

func (a *EthereumAddressAgentID) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(a.Kind()))
	ww.Write(&a.chainID)
	ww.WriteN(a.eth[:])
	return ww.Err
}
