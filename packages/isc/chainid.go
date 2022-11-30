// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/parameters"
)

const ChainIDLength = iotago.AliasIDLength

// ChainID represents the global identifier of the chain
// It is wrapped AliasAddress, an address without a private key behind
type ChainID iotago.AliasID

// ChainIDFromAliasID creates new chain ID from alias address
func ChainIDFromAliasID(addr iotago.AliasID) ChainID {
	return ChainID(addr)
}

// ChainIDFromBytes reconstructs a ChainID from its binary representation.
func ChainIDFromBytes(data []byte) (*ChainID, error) {
	var ret ChainID
	if len(ret) != len(data) {
		return &ChainID{}, xerrors.New("cannot decode ChainID: wrong data length")
	}
	copy(ret[:], data)
	return &ret, nil
}

func ChainIDFromString(s string) (*ChainID, error) {
	_, addr, err := iotago.ParseBech32(s)
	if err != nil {
		return nil, err
	}
	aliasAddr, ok := addr.(*iotago.AliasAddress)
	if !ok {
		return nil, fmt.Errorf("chainID must be an alias address")
	}
	cid := ChainIDFromAddress(aliasAddr)
	return &cid, nil
}

// ChainIDFromMarshalUtil reads from Marshalutil
func ChainIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ChainID, error) {
	bin, err := mu.ReadBytes(ChainIDLength)
	if err != nil {
		return &ChainID{}, err
	}
	return ChainIDFromBytes(bin)
}

func ChainIDFromAddress(addr *iotago.AliasAddress) ChainID {
	var alias iotago.AliasID
	copy(alias[:], addr[:])
	return ChainIDFromAliasID(alias)
}

// RandomChainID creates a random chain ID. Used for testing only
func RandomChainID(seed ...[]byte) *ChainID {
	var h hashing.HashValue
	if len(seed) > 0 {
		h = hashing.HashData(seed[0])
	} else {
		h = hashing.RandomHash(nil)
	}
	ret, _ := ChainIDFromBytes(h[:ChainIDLength])
	return ret
}

func (chid *ChainID) AsAliasID() *iotago.AliasID {
	return (*iotago.AliasID)(chid)
}

func (chid *ChainID) Bytes() []byte {
	return chid[:]
}

func (chid *ChainID) Key() string {
	return chid.AsAliasID().String()
}

// Equals for using
func (chid *ChainID) Equals(chid1 *ChainID) bool {
	if chid == chid1 {
		return true
	}
	if chid == nil || chid1 == nil {
		return false
	}
	return *chid == *chid1
}

// String human readable form (bech32)
func (chid *ChainID) String() string {
	return chid.AsAddress().Bech32(parameters.L1().Protocol.Bech32HRP)
}

func (chid *ChainID) ShortString() string {
	return chid.AsAddress().String()[2:8]
}

func (chid *ChainID) AsAddress() iotago.Address {
	ret := iotago.AliasAddress(*chid)
	return &ret
}

func (chid *ChainID) AsAliasAddress() *iotago.AliasAddress {
	ret := iotago.AliasAddress(*chid)
	return &ret
}

func (chid *ChainID) CommonAccount() AgentID {
	return NewContractAgentID(chid, 0)
}

func (chid *ChainID) IsSameChain(aid AgentID) bool {
	contract, ok := aid.(*ContractAgentID)
	if !ok {
		return false
	}
	return chid.Equals(contract.ChainID())
}
