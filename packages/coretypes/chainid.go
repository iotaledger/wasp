// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"io"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/mr-tron/base58"
)

// ChainIDLength size of the ChainID in bytes
const ChainIDLength = address.Length

// ChainID represents the global identifier of the chain
//
// Currently it is an alias for the chain address (type address.Address)
// In the future it will be refactored as an alias for chain color (type balance.Color)
type ChainID address.Address

var NilChainID = ChainID{}

// NewChainIDFromBase58 constructor decodes base58 string to the ChainID
func NewChainIDFromBase58(b58 string) (ret ChainID, err error) {
	var b []byte
	b, err = base58.Decode(b58)
	if err != nil {
		return
	}
	if len(b) != ChainIDLength {
		err = ErrWrongDataLength
		return
	}
	copy(ret[:], b)
	return
}

// NewChainIDFromBytes reconstructs a ChainID from its binary representation.
func NewChainIDFromBytes(data []byte) (ret ChainID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewRandomChainID creates a random chain ID.
func NewRandomChainID() ChainID {
	return ChainID(address.RandomOfType(address.VersionBLS))
}

// Bytes returns the ChainID as byte slice.
func (chid ChainID) Bytes() []byte {
	return address.Address(chid).Bytes()
}

// String human readable form (base58 encoding)
func (chid ChainID) String() string {
	return address.Address(chid).String()
}

// Write to writer
func (chid *ChainID) Write(w io.Writer) error {
	_, err := w.Write(chid[:])
	return err
}

// Read from reader
func (chid *ChainID) Read(r io.Reader) error {
	n, err := r.Read(chid[:])
	if err != nil {
		return err
	}
	if n != ChainIDLength {
		return ErrWrongDataLength
	}
	return nil
}
