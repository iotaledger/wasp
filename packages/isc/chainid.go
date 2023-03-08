// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/parameters"
)

const ChainIDLength = iotago.AliasIDLength

var emptyChainID = ChainID{}

// ChainID represents the global identifier of the chain
// It is wrapped AliasAddress, an address without a private key behind
type ChainID iotago.AliasID

// EmptyChainID returns an empty ChainID.
func EmptyChainID() ChainID {
	return emptyChainID
}

// ChainIDFromAliasID creates new chain ID from alias address
func ChainIDFromAliasID(aliasID iotago.AliasID) ChainID {
	return ChainID(aliasID)
}

// ChainIDFromBytes reconstructs a ChainID from its binary representation.
func ChainIDFromBytes(data []byte) (ChainID, error) {
	var chainID ChainID
	if ChainIDLength != len(data) {
		return ChainID{}, errors.New("cannot decode ChainID: wrong data length")
	}
	copy(chainID[:], data)
	return chainID, nil
}

func ChainIDFromString(bech32 string) (ChainID, error) {
	_, addr, err := iotago.ParseBech32(bech32)
	if err != nil {
		return ChainID{}, err
	}
	if addr.Type() != iotago.AddressAlias {
		return ChainID{}, fmt.Errorf("chainID must be an alias address (%s)", bech32)
	}
	return ChainIDFromAddress(addr.(*iotago.AliasAddress)), nil
}

// ChainIDFromMarshalUtil reads from Marshalutil
func ChainIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (ChainID, error) {
	bin, err := mu.ReadBytes(ChainIDLength)
	if err != nil {
		return ChainID{}, err
	}
	return ChainIDFromBytes(bin)
}

func ChainIDFromAddress(addr *iotago.AliasAddress) ChainID {
	return ChainIDFromAliasID(addr.AliasID())
}

// RandomChainID creates a random chain ID. Used for testing only
func RandomChainID(seed ...[]byte) ChainID {
	var h hashing.HashValue
	if len(seed) > 0 {
		h = hashing.HashData(seed[0])
	} else {
		h = hashing.PseudoRandomHash(nil)
	}
	chainID, err := ChainIDFromBytes(h[:ChainIDLength])
	if err != nil {
		panic(err)
	}
	return chainID
}

func (id ChainID) AsAliasID() iotago.AliasID {
	return iotago.AliasID(id)
}

func (id ChainID) Bytes() []byte {
	return id[:]
}

func (id ChainID) Key() string {
	return id.AsAliasID().String()
}

func (id ChainID) Equals(other ChainID) bool {
	return id == other
}

func (id ChainID) Empty() bool {
	return id == emptyChainID
}

// String human readable form (bech32)
func (id ChainID) String() string {
	return id.AsAddress().Bech32(parameters.L1().Protocol.Bech32HRP)
}

func (id ChainID) ShortString() string {
	return id.AsAddress().String()[0:10]
}

func (id ChainID) AsAddress() iotago.Address {
	addr := iotago.AliasAddress(id)
	return &addr
}

func (id ChainID) AsAliasAddress() iotago.AliasAddress {
	return iotago.AliasAddress(id)
}

func (id ChainID) IsSameChain(agentID AgentID) bool {
	contract, ok := agentID.(*ContractAgentID)
	if !ok {
		return false
	}
	return id.Equals(contract.ChainID())
}
