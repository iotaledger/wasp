// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/util"
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

func NewAgentIDFromBytes(data []byte) (ret AgentID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
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

func (a *AgentID) Bytes() []byte {
	var buf bytes.Buffer
	_ = a.Write(&buf)
	return buf.Bytes()
}

// IsAddress checks if agentID represents address. 0 in the place of the contract's hname means it is an address
// This is based on the assumption that fro coretypes.Hname 0 is a reserved value
func (a *AgentID) IsAddress() bool {
	switch a.a.(type) {
	case *ContractID:
		return false
	case ledgerstate.Address:
		return true
	}
	panic("wrong type behind AgentID")
}

// MustAddress takes address or panic if not address
func (a *AgentID) MustAddress() ledgerstate.Address {
	return a.a.(ledgerstate.Address)
}

// MustContractID takes contract ID or panics if not a contract ID
func (a *AgentID) MustContractID() *ContractID {
	return a.a.(*ContractID)
}

// String human readable string
func (a *AgentID) String() string {
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

func (a *AgentID) Write(w io.Writer) error {
	if err := util.WriteBoolByte(w, a.IsAddress()); err != nil {
		return err
	}
	if a.IsAddress() {
		if _, err := w.Write(a.MustAddress().Bytes()); err != nil {
			return err
		}
	} else {
		if err := a.MustContractID().Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (a *AgentID) Read(r io.Reader) error {
	var isAddress bool
	if err := util.ReadBoolByte(r, &isAddress); err != nil {
		return err
	}
	if isAddress {
		var data [ledgerstate.AddressLength]byte
		if n, err := r.Read(data[:]); err != nil || n != ledgerstate.AddressLength {
			return ErrWrongDataLength
		}
		t, _, err := ledgerstate.AddressFromBytes(data[:])
		if err != nil {
			return xerrors.Errorf("failed parsing address: %v", err)
		}
		a.a = t
		return nil
	}
	cid := new(ContractID)
	if err := cid.Read(r); err != nil {
		return err
	}
	a.a = cid
	return nil
}
