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
func NewContractID(chid ChainID, contractHn Hname) (ret ContractID) {
	return ContractID{
		chainID:       chid,
		contractHname: contractHn,
	}
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

func (scid ContractID) Clone() ContractID {
	return scid
}

// ChainID returns ID of the native chain of the contract
func (scid ContractID) ChainID() ChainID {
	return scid.chainID
}

// Hname returns hashed name of the contract, local ID on the chain
func (scid ContractID) Hname() Hname {
	return scid.contractHname
}

// Base58 base58 representation of the binary representation
func (scid ContractID) Base58() string {
	var buf bytes.Buffer
	buf.Write(scid.chainID[:])
	buf.Write(scid.contractHname.Bytes())
	return base58.Encode(buf.Bytes())
}

const (
	long_format  = "%s::%s"
	short_format = "%s..::%s"
)

// String human readable representation of the contract ID <chainID>::<hanme>
func (scid *ContractID) String() string {
	return fmt.Sprintf(long_format, scid.ChainID().String(), scid.Hname().String())
}

// Short human readable representation in short form
func (scid *ContractID) Short() string {
	return fmt.Sprintf(short_format, scid.ChainID().String()[:8], scid.Hname().String())
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
	if _, err := w.Write(scid.chainID[:]); err != nil {
		return err
	}
	if _, err := w.Write(scid.contractHname.Bytes()); err != nil {
		return err
	}
	return nil
}
