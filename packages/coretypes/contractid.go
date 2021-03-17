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

// ContractID global identifier of the smart contract.
// It consists of chainID and the hname of the contract on the chain
type ContractID struct {
	chainID       ChainID
	contractHname Hname
}

// NewContractID creates new ContractID from chainID and contract hname
func NewContractID(chid ChainID, contractHn Hname) *ContractID {
	return &ContractID{
		chainID:       chid,
		contractHname: contractHn,
	}
}

// NewContractIDFromBytes creates contract ID frm its binary representation
func NewContractIDFromBytes(data []byte) (*ContractID, error) {
	ret := ContractID{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return &ret, nil
}

// NewContractIDFromBase58 decodes contract ID from base58 string
func NewContractIDFromBase58(base58string string) (*ContractID, error) {
	data, err := base58.Decode(base58string)
	if err != nil {
		return nil, err
	}
	return NewContractIDFromBytes(data)
}

// NewContractIDFromString parses the human-readable string representation of the contract ID
func NewContractIDFromString(s string) (*ContractID, error) {
	parts := strings.Split(s, "::")
	if len(parts) != 2 {
		return nil, errors.New("invalid ContractID")
	}
	chid, err := NewChainIDFromBase58(parts[0])
	if err != nil {
		return nil, err
	}
	cid, err := HnameFromString(parts[1])
	if err != nil {
		return nil, err
	}
	return NewContractID(chid, cid), nil
}

func (scid *ContractID) Clone() *ContractID {
	return &ContractID{
		chainID:       scid.chainID.Clone(),
		contractHname: scid.contractHname,
	}
}

func (scid *ContractID) Equals(scid1 *ContractID) bool {
	return scid.chainID.Equals(&scid1.chainID) && scid.contractHname == scid1.contractHname
}

// ChainID returns ID of the native chain of the contract
func (scid *ContractID) ChainID() *ChainID {
	return &scid.chainID
}

// Hname returns hashed name of the contract, local ID on the chain
func (scid *ContractID) Hname() Hname {
	return scid.contractHname
}

// Base58 base58 representation of the binary representation
func (scid *ContractID) Base58() string {
	return base58.Encode(scid.Bytes())
}

const (
	long_format  = "%s::%s"
	short_format = "%s..::%s"
)

// String human readable representation of the contract ID <chainID>::<hanme>
func (scid *ContractID) String() string {
	return fmt.Sprintf(long_format, scid.ChainID().Base58(), scid.Hname().String())
}

// Short human readable representation in short form
func (scid *ContractID) Short() string {
	return fmt.Sprintf(short_format, scid.ChainID().Base58()[:8], scid.Hname().String())
}

func (scid *ContractID) Bytes() []byte {
	var buf bytes.Buffer
	_ = scid.Write(&buf)
	return buf.Bytes()
}

// Read from reader
func (scid *ContractID) Read(r io.Reader) error {
	if err := scid.chainID.Read(r); err != nil {
		return err
	}
	if err := scid.contractHname.Read(r); err != nil {
		return err
	}
	return nil
}

// Write to writer
func (scid *ContractID) Write(w io.Writer) error {
	if _, err := w.Write(scid.chainID.Bytes()); err != nil {
		return err
	}
	if _, err := w.Write(scid.contractHname.Bytes()); err != nil {
		return err
	}
	return nil
}
