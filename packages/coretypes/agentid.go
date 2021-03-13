// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"golang.org/x/xerrors"
	"io"
)

// AgentID represents exactly one of two types of entities on the ISCP ledger in one ID:
//  - It can represent an address on the Tangle (controlled by some private key). In this case it can be
//    interpreted as address.Address type (see MustAddress).
//  - alternatively, it can represent a smart contract on the ISCP. In this case it can be interpreted as
//    a coretypes.ContractID type (see MustContractID)
// Type of ID represented by the AgentID can be recognized with IsAddress call.
// An attempt to interpret the AgentID in the wrong way invokes panic
type AgentID struct {
	a interface{} // one of 2: *ContractID or ledgerstate.Address
}

// NewAgentIDFromContractID makes AgentID from ContractID
func NewAgentIDFromContractID(id ContractID) AgentID {
	ret := id.Clone()
	return AgentID{&ret}
}

// NewAgentIDFromAddress makes AgentID from address.Address
func NewAgentIDFromAddress(addr ledgerstate.Address) (AgentID, error) {
	if addr.Type() == ledgerstate.AliasAddressType {
		return AgentID{}, xerrors.New("AgentID cannot be based on AliasAddress")
	}
	return AgentID{addr.Clone()}, nil
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() AgentID {
	chainID := NewRandomChainID()
	hname := Hn("testFunction")
	return NewAgentIDFromContractID(NewContractID(chainID, hname))
}

// IsAddress checks if agentID represents address. 0 in the place of the contract's hname means it is an address
// This is based on the assumption that fro coretypes.Hname 0 is a reserved value
func (a AgentID) IsAddress() bool {
	switch a.a.(type) {
	case *ContractID:
		return false
	case ledgerstate.Address:
		return true
	}
	panic("wrong type behind AgentID")
}

// MustAddress takes address or panic if not address
func (a AgentID) MustAddress() ledgerstate.Address {
	return a.a.(ledgerstate.Address)
}

// MustContractID takes contract ID or panics if not a contract ID
func (a AgentID) MustContractID() ContractID {
	return *a.a.(*ContractID)
}

// String human readable string
func (a AgentID) String() string {
	if a.IsAddress() {
		return "A/" + a.MustAddress().Base58()
	}
	cid := a.MustContractID()
	return "C/" + cid.String()
}

// NewAgentIDFromString parses the human-readable string representation
func NewAgentIDFromString(s string) (ret AgentID, err error) {
	if len(s) < 2 {
		err = errors.New("invalid length")
		return
	}
	switch s[:2] {
	case "A/":
		var addr ledgerstate.Address
		addr, err = ledgerstate.AddressFromBase58EncodedString(s[2:])
		if err != nil {
			return
		}
		ret, err = NewAgentIDFromAddress(addr)
		return
	case "C/":
		var cid ContractID
		cid, err = NewContractIDFromString(s[2:])
		if err != nil {
			return
		}
		return NewAgentIDFromContractID(cid), nil
	default:
		err = errors.New("invalid prefix")
	}
	return
}

// ReadAgentID decodes from binary representation
func ReadAgentID(r io.Reader, agentID *AgentID) error {
	var b [1]byte
	if n, err := r.Read(b[:]); err != nil || n != 1 {
		return ErrWrongDataLength
	}
	switch b[0] {
	case 'A':
		var a [ledgerstate.AddressLength]byte
		if n, err := r.Read(a[:]); err != nil || n != ledgerstate.AddressLength {
			return ErrWrongDataLength
		}
		addr, _, err := ledgerstate.AddressFromBytes(a[:])
		if err != nil {
			return err
		}
		agentID.a = addr
	case 'C':
		var cid ContractID
		if err := cid.Read(r); err != nil {
			return err
		}
		agentID.a = &cid
	default:
		return xerrors.New("error while reading AgentID")
	}
	return nil
}
