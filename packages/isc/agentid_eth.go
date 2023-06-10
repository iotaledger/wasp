package isc

import (
	"errors"
	"io"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// EthereumAddressAgentID is an AgentID formed by an Ethereum address
type EthereumAddressAgentID struct {
	eth common.Address
}

var _ AgentID = &EthereumAddressAgentID{}

func NewEthereumAddressAgentID(eth common.Address) *EthereumAddressAgentID {
	return &EthereumAddressAgentID{eth: eth}
}

func ethAgentIDFromString(s string) (AgentID, error) {
	eth := common.HexToAddress(s)
	return NewEthereumAddressAgentID(eth), nil
}

func (a *EthereumAddressAgentID) EthAddress() common.Address {
	return a.eth
}

func (a *EthereumAddressAgentID) Kind() rwutil.Kind {
	return AgentIDKindEthereumAddress
}

func (a *EthereumAddressAgentID) Bytes() []byte {
	return rwutil.WriterToBytes(a)
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
	return other.(*EthereumAddressAgentID).eth == a.eth
}

func (a *EthereumAddressAgentID) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	kind := rwutil.Kind(rr.ReadByte())
	if rr.Err == nil && kind != a.Kind() {
		return errors.New("invalid EthereumAddressAgentID kind")
	}
	rr.ReadN(a.eth[:])
	return rr.Err
}

func (a *EthereumAddressAgentID) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint8(uint8(a.Kind()))
	ww.WriteN(a.eth[:])
	return ww.Err
}
