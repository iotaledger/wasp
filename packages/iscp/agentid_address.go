package iscp

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
)

// AddressAgentID is an AgentID backed by a non-alias address.
type AddressAgentID struct {
	a iotago.Address
}

var _ AgentIDWithL1Address = &AddressAgentID{}

func addressAgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (AgentID, error) {
	var addr iotago.Address
	var err error
	if addr, err = AddressFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return NewAgentID(addr), nil
}

func addressAgentIDFromString(s string, networkPrefix iotago.NetworkPrefix) (AgentID, error) {
	_, addr, err := iotago.ParseBech32(s)
	if err != nil {
		return nil, err
	}
	return NewAgentID(addr), nil
}

func (a *AddressAgentID) Address() iotago.Address {
	return a.a
}

func (a *AddressAgentID) Kind() AgentIDKind {
	return AgentIDKindAddress
}

func (a *AddressAgentID) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(byte(a.Kind()))
	addressBytes, err := a.a.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(err)
	}
	mu.WriteBytes(addressBytes)
	return mu.Bytes()
}

func (a *AddressAgentID) String() string {
	return a.a.Bech32(parameters.L1.Protocol.Bech32HRP)
}

func (a *AddressAgentID) Equals(other AgentID) bool {
	if other == nil {
		return false
	}
	if other.Kind() != a.Kind() {
		return false
	}
	o := other.(*AddressAgentID)
	return o.a.Equal(a.a)
}
