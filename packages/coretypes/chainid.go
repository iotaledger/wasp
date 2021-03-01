// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
	"io"

	"github.com/iotaledger/hive.go/crypto/ed25519"
)

// ChainIDLength size of the ChainID in bytes
const ChainIDLength = ledgerstate.AddressLength

// ChainID represents the global identifier of the chain
//
// Currently it is an alias for the chain address (type address.Address)
// In the future it will be refactored as an alias for chain color (type balance.Color)
type ChainID [ChainIDLength]byte

var NilChainID = ChainID{}

// NewChainIDFromAddress temporary
func NewChainIDFromAddress(addr ledgerstate.Address) (ret ChainID) {
	copy(ret[:], addr.Bytes())
	return
}

// NewChainIDFromBase58 constructor decodes base58 string to the ChainID
func NewChainIDFromBase58(b58 string) (ret ChainID, err error) {
	addr, err := ledgerstate.AddressFromBase58EncodedString(b58)
	if err != nil {
		return
	}
	copy(ret[:], addr.Bytes())
	return
}

// NewChainIDFromBytes reconstructs a ChainID from its binary representation.
func NewChainIDFromBytes(data []byte) (ret ChainID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewRandomChainID creates a random chain ID.
func NewRandomChainID(seed ...[]byte) (ret ChainID) {
	var h hashing.HashValue
	if len(seed) > 0 {
		h = hashing.HashData(seed[0])
	} else {
		h = hashing.RandomHash(nil)
	}
	copy(ret[:], ledgerstate.NewED25519Address(ed25519.NewSeed(h[:]).KeyPair(0).PublicKey).Bytes())
	return
}

func (chid ChainID) Clone() ChainID {
	return chid
}

func (chid ChainID) Base58() string {
	return base58.Encode(chid[:])
}

// String human readable form (base58 encoding)
func (chid ChainID) String() string {
	return "$/" + chid.Base58()
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

// AsAddress Temporary
func (chid *ChainID) AsAddress() ledgerstate.Address {
	ret, _, err := ledgerstate.AddressFromBytes(chid[:])
	if err != nil {
		panic(xerrors.Errorf("ChainID.MustAsAddress: %v", err))
	}
	return ret
}
