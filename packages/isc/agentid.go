// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type AgentIDKind byte

const (
	AgentIDKindNil AgentIDKind = iota
	AgentIDKindAddress
	AgentIDKindContract
	AgentIDKindEthereumAddress

	AgentIDIsNil AgentIDKind = 0x80
)

// AgentID represents any entity that can hold assets on L2 and/or call contracts.
type AgentID interface {
	Bytes() []byte
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

	addressAsBytes, err := hexutil.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("malformed AgentID '%v' %w", s, err)
	}

	// HName
	if len(addressAsBytes) == HnameLength {
		return contractAgentIDFromString(s)
	}

	// EVM
	if len(addressAsBytes) == common.AddressLength {
		return ethAgentIDFromString(s)
	}

	// Rebased
	if len(addressAsBytes) == iotago.AddressLen {
		return addressAgentIDFromString(s)
	}

	return nil, errors.Errorf("AgentIDFromString: invalid address length %d", len(addressAsBytes))
}
