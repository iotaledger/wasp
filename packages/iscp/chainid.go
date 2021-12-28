// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"encoding/hex"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
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

func ChainIDFromHex(s string) (*ChainID, error) {
	bin, err := hex.DecodeString(s)
	if err != nil {
		return &ChainID{}, err
	}
	return ChainIDFromBytes(bin)
}

// TODO adjust to iotago style
// ChainIDFromMarshalUtil reads from Marshalutil
func ChainIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ChainID, error) {
	bin, err := mu.ReadBytes(ChainIDLength)
	if err != nil {
		return &ChainID{}, err
	}
	return ChainIDFromBytes(bin)
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

// Equals for using
func (chid *ChainID) Equals(chid1 *ChainID) bool {
	return chid == chid1
}

// String human readable form (hex encoding)
func (chid *ChainID) String() string {
	return "$/" + chid.Hex()
}

func (chid *ChainID) Hex() string {
	return hex.EncodeToString(chid[:])
}

func (chid *ChainID) AsAddress() iotago.Address {
	ret := iotago.AliasAddress(*chid)
	return &ret
}

func (chid *ChainID) AsAliasAddress() iotago.AliasAddress {
	return iotago.AliasAddress(*chid)
}
