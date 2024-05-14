package isc

import (
	"io"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// AddressAgentID is an AgentID backed by a non-alias address.
type AddressAgentID struct {
	a cryptolib.Address
}

var _ AgentIDWithL1Address = &AddressAgentID{}

func NewAddressAgentID(addr cryptolib.Address) *AddressAgentID {
	return &AddressAgentID{a: addr}
}

func addressAgentIDFromString(s string) (*AddressAgentID, error) {
	_, addr, err := cryptolib.AddressFromBech32(s)
	if err != nil {
		return nil, err
	}
	return &AddressAgentID{a: addr}, nil
}

func (a *AddressAgentID) Address() cryptolib.Address {
	return a.a
}

func (a *AddressAgentID) Bytes() []byte {
	return rwutil.WriteToBytes(a)
}

func (a *AddressAgentID) BelongsToChain(ChainID) bool {
	return false
}

func (a *AddressAgentID) BytesWithoutChainID() []byte {
	return a.Bytes()
}

func (a *AddressAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	if other.Kind() != a.Kind() {
		return false
	}
	return other.(*AddressAgentID).a.Equal(a.a)
}

func (a *AddressAgentID) Kind() AgentIDKind {
	return AgentIDKindAddress
}

func (a *AddressAgentID) String() string {
	return cryptolib.AddressToBech32String(parameters.L1().Protocol.Bech32HRP, a.a)
}

func (a *AddressAgentID) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(rwutil.Kind(a.Kind()))
	a.a = cryptolib.ReadAddress(rr)
	return rr.Err
}

func (a *AddressAgentID) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(a.Kind()))
	cryptolib.WriteAddress(a.a, ww)
	return ww.Err
}
