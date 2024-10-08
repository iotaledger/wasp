// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
)

const ChainIDLength = sui.AddressLen

var emptyChainID = ChainID{}

// ChainID represents the global identifier of the chain
// It is wrapped AliasAddress, an address without a private key behind
type (
	// ChainID is the anchor ObjectID
	ChainID    sui.ObjectID
	ChainIDKey string
)

// EmptyChainID returns an empty ChainID.
func EmptyChainID() ChainID {
	return emptyChainID
}

func ChainIDFromAddress(addr *cryptolib.Address) ChainID {
	return ChainID(addr[:])
}

func ChainIDFromObjectID(addr sui.ObjectID) ChainID {
	return ChainID(addr[:])
}

// ChainIDFromBytes reconstructs a ChainID from its binary representation.
func ChainIDFromBytes(data []byte) (ret ChainID, err error) {
	return bcs.Unmarshal[ChainID](data)
}

func ChainIDFromString(hexAddress string) (ChainID, error) {
	addr, err := cryptolib.NewAddressFromHexString(hexAddress)
	if err != nil {
		return ChainID{}, err
	}

	return ChainIDFromAddress(addr), nil
}

func ChainIDFromKey(key ChainIDKey) ChainID {
	chainID, err := ChainIDFromString(string(key))
	if err != nil {
		panic(err)
	}
	return chainID
}

func (id ChainID) AsObjectID() sui.ObjectID {
	return sui.ObjectID(id)
}

func (id ChainID) AsAddress() *cryptolib.Address {
	addr := cryptolib.Address(id)
	return &addr
}

func (id ChainID) AsAliasAddress() iotago.AliasAddress {
	return iotago.AliasAddress(id)
}

func (id ChainID) AsAliasID() iotago.AliasID {
	return iotago.AliasID(id)
}

func (id ChainID) Bytes() []byte {
	return id[:]
}

func (id ChainID) Empty() bool {
	return id == emptyChainID
}

func (id ChainID) Equals(other ChainID) bool {
	return id == other
}

func (id ChainID) Key() ChainIDKey {
	return ChainIDKey(id.AsAliasID().String())
}

func (id ChainID) IsSameChain(agentID AgentID) bool {
	contract, ok := agentID.(*ContractAgentID)
	if !ok {
		return false
	}
	return id.Equals(contract.ChainID())
}

func (id ChainID) ShortString() string {
	return id.AsAddress().String()[0:10]
}

func (id ChainID) String() string {
	return id.AsAddress().String()
}
