// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"errors"
	"strings"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type AgentIDKind byte

const (
	AgentIDKindNil AgentIDKind = iota
	AgentIDKindAddress
	AgentIDKindContract
	AgentIDKindEthereumAddress

	AgentIDIsNil AgentIDKind = 0x80
)

const AgentIDStringSeparator = "@"

// AgentID represents any entity that can hold assets on L2 and/or call contracts.
type AgentID interface {
	Bytes() []byte
	BelongsToChain(ChainID) bool
	BytesWithoutChainID() []byte
	Equals(other AgentID) bool
	Kind() AgentIDKind
	String() string
}

func init() {
	bcs.RegisterEnumType5[AgentID, bcs.None, *AddressAgentID, *ContractAgentID, *EthereumAddressAgentID, *NilAgentID]()
}

// AgentIDWithL1Address is an AgentID backed by an L1 address (either AddressAgentID or ContractAgentID).
type AgentIDWithL1Address interface {
	AgentID
	Address() *cryptolib.Address
}

// AddressFromAgentID returns the L1 address of the AgentID, if applicable.
func AddressFromAgentID(a AgentID) (*cryptolib.Address, bool) {
	wa, ok := a.(AgentIDWithL1Address)
	if !ok {
		return nil, false
	}
	return wa.Address(), true
}

// HnameFromAgentID returns the hname of the AgentID, or HnameNil if not applicable.
func HnameFromAgentID(a AgentID) Hname {
	if ca, ok := a.(*ContractAgentID); ok {
		return ca.Hname()
	}
	return HnameNil
}

func AgentIDFromBytes(data []byte) (AgentID, error) {
	return bcs.Unmarshal[AgentID](data)
}

// AgentIDFromString parses the human-readable string representation
func AgentIDFromString(s string) (AgentID, error) {
	if s == nilAgentIDString {
		return &NilAgentID{}, nil
	}
	var contractPart, addrPart string
	{
		parts := strings.Split(s, AgentIDStringSeparator)
		switch len(parts) {
		case 1:
			addrPart = parts[0]
		case 2:
			addrPart = parts[1]
			contractPart = parts[0]
		default:
			return nil, errors.New("invalid AgentID format")
		}
	}

	if contractPart != "" {
		if strings.HasPrefix(contractPart, "0x") {
			return ethAgentIDFromString(contractPart, addrPart)
		}
		return contractAgentIDFromString(contractPart, addrPart)
	}

	return addressAgentIDFromString(s)
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() AgentID {
	return NewContractAgentID(RandomChainID(), Hn("testName"))
}
