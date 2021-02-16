// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mr-tron/base58"
)

// ContractIDLength size of the contract ID in bytes
const ContractIDLength = ChainIDLength + HnameLength

// ContractID global identifier of the smart contract. It consists of chainID and the hname of the contract on the chain
type ContractID [ContractIDLength]byte

// NewContractID creates new ContractID from chainID and contract hname
func NewContractID(chid ChainID, contractHn Hname) (ret ContractID) {
	copy(ret[:ChainIDLength], chid[:])
	copy(ret[ChainIDLength:], contractHn.Bytes())
	return
}

// NewContractIDFromBytes creates contract ID frm its binary representation
func NewContractIDFromBytes(data []byte) (ret ContractID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewContractIDFromBase58 decodes contract ID from base58 string
func NewContractIDFromBase58(base58string string) (ret ContractID, err error) {
	var data []byte
	if data, err = base58.Decode(base58string); err != nil {
		return
	}
	return NewContractIDFromBytes(data)
}

// NewContractIDFromString parses the human-readable string representation of the contract ID
func NewContractIDFromString(s string) (ret ContractID, err error) {
	parts := strings.Split(s, "::")
	if len(parts) != 2 {
		err = errors.New("invalid ContractID")
		return
	}
	chid, err := NewChainIDFromBase58(parts[0])
	if err != nil {
		return
	}
	cid, err := HnameFromString(parts[1])
	if err != nil {
		return
	}
	ret = NewContractID(chid, cid)
	return
}

// ChainID returns ID of the native chain of the contract
func (scid ContractID) ChainID() (ret ChainID) {
	copy(ret[:ChainIDLength], scid[:ChainIDLength])
	return
}

// Hname returns hashed name of the contract, local ID on the chain
func (scid ContractID) Hname() Hname {
	ret, _ := NewHnameFromBytes(scid[ChainIDLength:])
	return ret
}

// Base58 base58 representation of the binary representation
func (scid ContractID) Base58() string {
	return base58.Encode(scid[:])
}

const (
	long_format  = "%s::%s"
	short_format = "%s..::%s"
)

// Bytes contract ID as byte slice
func (scid ContractID) Bytes() []byte {
	return scid[:]
}

// String human readable representation of the contract ID <chainID>::<hanme>
func (scid ContractID) String() string {
	return fmt.Sprintf(long_format, scid.ChainID().String(), scid.Hname().String())
}

// Short human readable representation in short form
func (scid ContractID) Short() string {
	return fmt.Sprintf(short_format, scid.ChainID().String()[:8], scid.Hname().String())
}

// Read from reated
func (scid *ContractID) Read(r io.Reader) error {
	n, err := r.Read(scid[:])
	if err != nil {
		return err
	}
	if n != ContractIDLength {
		return ErrWrongDataLength
	}
	return nil
}

// Write to writer
func (scid *ContractID) Write(w io.Writer) error {
	_, err := w.Write(scid[:])
	return err
}
