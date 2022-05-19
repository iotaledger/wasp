// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"strings"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
	"golang.org/x/xerrors"
)

type AgentIDKind uint8

const (
	AgentIDKindNil AgentIDKind = iota
	AgentIDKindAddress
	AgentIDKindContract
	AgentIDKindEthereumAddress
)

// AgentID represents any entity that can hold assets on L2 and/or call contracts.
type AgentID interface {
	Kind() AgentIDKind

	String() string
	Bytes() []byte
	Equals(other AgentID) bool
}

// AgentIDWithL1Address is an AgentID backed by an L1 address (either AddressAgentID or ContractAgentID).
type AgentIDWithL1Address interface {
	AgentID
	Address() iotago.Address
}

// AddressFromAgentID returns the L1 address of the AgentID, if applicable.
func AddressFromAgentID(a AgentID) (iotago.Address, bool) {
	wa, ok := a.(AgentIDWithL1Address)
	if !ok {
		return nil, false
	}
	return wa.Address(), true
}

// HnameFromAgentID returns the hname of the AgentID, if applicable.
func HnameFromAgentID(a AgentID) (Hname, bool) {
	ca, ok := a.(*ContractAgentID)
	if !ok {
		return 0, false
	}
	return ca.Hname(), true
}

// NewAgentID creates an AddressAgentID if the address is not an AliasAddress;
// otherwise a ContractAgentID with hname = 0.
func NewAgentID(addr iotago.Address) AgentID {
	if addr.Type() == iotago.AddressAlias {
		chid := ChainIDFromAddress(addr.(*iotago.AliasAddress))
		return NewContractAgentID(&chid, 0)
	}
	return &AddressAgentID{a: addr}
}

// Deprecated: NewAgentIDFromAddressAndHname is a utility constructor for backwards compatibility.
// Use NewAgentID / NewContractAgentID / NewEthereumAddressAgentID instead.
func NewAgentIDFromAddressAndHname(addr iotago.Address, hname Hname) AgentID {
	if hname != 0 {
		if addr.Type() != iotago.AddressAlias {
			panic("inconsistency: non-alias address cannot have hname != 0")
		}
		chid := ChainIDFromAddress(addr.(*iotago.AliasAddress))
		return NewContractAgentID(&chid, hname)
	}
	return NewAgentID(addr)
}

func AgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (AgentID, error) {
	var err error
	kind, err := mu.ReadByte()
	if err != nil {
		return nil, err
	}
	switch AgentIDKind(kind) {
	case AgentIDKindNil:
		return &NilAgentID{}, nil
	case AgentIDKindAddress:
		return addressAgentIDFromMarshalUtil(mu)
	case AgentIDKindContract:
		return contractAgentIDFromMarshalUtil(mu)
	case AgentIDKindEthereumAddress:
		return ethAgentIDFromMarshalUtil(mu)
	}
	return nil, xerrors.Errorf("no handler for AgentID kind %d", kind)
}

func AgentIDFromBytes(data []byte) (AgentID, error) {
	return AgentIDFromMarshalUtil(marshalutil.New(data))
}

// NewAgentIDFromString parses the human-readable string representation
func NewAgentIDFromString(s string) (AgentID, error) {
	if s == nilAgentIDString {
		return &NilAgentID{}, nil
	}
	var hnamePart, addrPart string
	{
		parts := strings.Split(s, "@")
		switch len(parts) {
		case 1:
			addrPart = parts[0]
		case 2:
			addrPart = parts[1]
			hnamePart = parts[0]
		default:
			return nil, xerrors.New("NewAgentIDFromString: wrong format")
		}
	}

	if hnamePart != "" {
		return contractAgentIDFromString(hnamePart, addrPart)
	}
	if strings.HasPrefix(addrPart, string(parameters.L1.Protocol.Bech32HRP)) {
		return addressAgentIDFromString(s)
	}
	if strings.HasPrefix(addrPart, "0x") {
		return ethAgentIDFromString(s)
	}
	return nil, xerrors.New("NewAgentIDFromString: wrong format")
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() AgentID {
	return NewContractAgentID(RandomChainID(), Hn("testName"))
}
