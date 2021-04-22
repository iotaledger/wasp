// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"io"
	"strings"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// AgentID represents address on the ledger with optional hname
// If address is and alias address and hname is != 0 the agent id is interpreted as
// ad contract id
type AgentID struct {
	a ledgerstate.Address
	h Hname
}

var NilAgentID AgentID

func init() {
	var b [ledgerstate.AddressLength]byte
	nilAddr, _, err := ledgerstate.AddressFromBytes(b[:])
	if err != nil {
		panic(err)
	}
	NilAgentID = AgentID{
		a: nilAddr,
		h: 0,
	}
}

// NewAgentID makes new AgentID
func NewAgentID(addr ledgerstate.Address, hname Hname) *AgentID {
	return &AgentID{
		a: addr.Clone(),
		h: hname,
	}
}

func NewAgentIDFromBytes(data []byte) (*AgentID, error) {
	ret := AgentID{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return &ret, nil
}

func NewAgentIDFromBase58EncodedString(s string) (*AgentID, error) {
	data, err := base58.Decode(s)
	if err != nil {
		return nil, err
	}
	return NewAgentIDFromBytes(data)
}

// NewAgentIDFromString parses the human-readable string representation
func NewAgentIDFromString(s string) (*AgentID, error) {
	if len(s) < 2 {
		return nil, xerrors.New("NewAgentIDFromString: invalid length")
	}
	if s[:2] != "A/" {
		return nil, xerrors.New("NewAgentIDFromString: wrong prefix")
	}
	parts := strings.Split(s[2:], "::")
	if len(parts) != 2 {
		return nil, xerrors.New("NewAgentIDFromString: wrong format")
	}
	addr, err := ledgerstate.AddressFromBase58EncodedString(parts[0])
	if err != nil {
		return nil, xerrors.Errorf("NewAgentIDFromString: %v", err)
	}
	hname, err := HnameFromString(parts[1])
	if err != nil {
		return nil, xerrors.Errorf("NewAgentIDFromString: %v", err)
	}
	return NewAgentID(addr, hname), nil

}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() *AgentID {
	addr := RandomChainID().AsAddress()
	hname := Hn("testName")
	return NewAgentID(addr, hname)
}

func (a *AgentID) Clone() *AgentID {
	return &AgentID{
		a: a.a.Clone(),
		h: a.h,
	}
}

func (a *AgentID) Address() ledgerstate.Address {
	return a.a
}

func (a *AgentID) Hname() Hname {
	return a.h
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

// String human readable string
func (a *AgentID) String() string {
	return "A/" + a.a.Base58() + "::" + a.h.String()
}

func (a *AgentID) Base58() string {
	return base58.Encode(a.Bytes())
}

func (a *AgentID) Write(w io.Writer) error {
	if a.a == nil {
		var t [ledgerstate.AddressLength]byte
		if _, err := w.Write(t[:]); err != nil {
			return err
		}
	} else {
		if _, err := w.Write(a.a.Bytes()); err != nil {
			return err
		}
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

func (a *AgentID) IsNil() bool {
	return a.Equals(&NilAgentID)
}
