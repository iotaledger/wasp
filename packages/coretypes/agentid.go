// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"golang.org/x/xerrors"
	"io"
)

// AgentID represents address on the ledger with optional hname
// If address is and alias address and hname is != 0 the agent id is interpreted as
// ad contract id
type AgentID struct {
	a ledgerstate.Address
	h Hname
}

func NewAgentIDFromBytes(data []byte) (ret AgentID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewAgentIDFromContractID makes AgentID from ContractID
func NewAgentIDFromContractID(id ContractID) AgentID {
	return AgentID{
		a: id.ChainID().AsAddress().Clone(),
		h: id.Hname(),
	}
}

// NewAgentIDFromAddress makes AgentID from address.Address
func NewAgentIDFromAddress(addr ledgerstate.Address) AgentID {
	return AgentID{
		a: addr.Clone(),
	}
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
		ret = NewAgentIDFromAddress(addr)
	case "C/":
		var cid ContractID
		cid, err = NewContractIDFromString(s[2:])
		if err != nil {
			return
		}
		ret = NewAgentIDFromContractID(cid)
	default:
		err = errors.New("invalid prefix")
	}
	return
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() AgentID {
	chainID := NewRandomChainID()
	hname := Hn("testFunction")
	cid := NewContractID(chainID, hname)
	return NewAgentIDFromContractID(cid)
}

func (a *AgentID) Clone() *AgentID {
	return &AgentID{
		a: a.a.Clone(),
		h: a.h,
	}
}

func (a *AgentID) Bytes() []byte {
	var buf bytes.Buffer
	_ = a.Write(&buf)
	return buf.Bytes()
}

func (a *AgentID) Equals(a1 *AgentID) bool {
	if !a.a.Equals(a1.a) {
		return false
	}
	if a.h != a1.h {
		return false
	}
	return true
}

// IsNonAliasAddress checks if agentID represents address. 0 in the place of the contract's hname means it is an address
// This is based on the assumption that fro coretypes.Hname 0 is a reserved value
func (a *AgentID) IsContract() bool {
	return a.a.Type() == ledgerstate.AliasAddressType && a.h != 0
}

// MustAddress takes address or panic if not address
func (a *AgentID) AsAddress() ledgerstate.Address {
	return a.a
}

// MustContractID takes contract ID or panics if not a contract ID
func (a *AgentID) MustContractID() ContractID {
	chainId, err := NewChainIDFromAddress(a.a)
	if err != nil {
		panic(err)
	}
	return NewContractID(chainId, a.h)
}

// String human readable string
func (a *AgentID) String() string {
	if !a.IsContract() {
		return "A/" + a.AsAddress().Base58()
	}
	cid := a.MustContractID()
	return "C/" + cid.String()
}

func (a *AgentID) Write(w io.Writer) error {
	if _, err := w.Write(a.a.Bytes()); err != nil {
		return err
	}
	if err := a.h.Write(w); err != nil {
		return err
	}
	return nil
}

func (a *AgentID) Read(r io.Reader) error {
	var buf [ledgerstate.AddressLength]byte
	if n, err := r.Read(buf[:]); err != nil || n != ledgerstate.AddressLength {
		return xerrors.Errorf("error while parsing address. Err: %v", err)
	}
	if t, _, err := ledgerstate.AddressFromBytes(buf[:]); err != nil {
		return err
	} else {
		a.a = t
	}
	if err := a.h.Read(r); err != nil {
		return err
	}
	return nil
}
