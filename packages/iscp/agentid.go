// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"strings"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
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
	if addr == nil {
		panic("NewAgentID: address can't be nil")
	}
	return &AgentID{
		a: addr.Clone(),
		h: hname,
	}
}

func AgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (*AgentID, error) {
	var err error
	ret := &AgentID{}
	if ret.a, err = ledgerstate.AddressFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.h, err = HnameFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func AgentIDFromBytes(data []byte) (*AgentID, error) {
	return AgentIDFromMarshalUtil(marshalutil.New(data))
}

func NewAgentIDFromBase58EncodedString(s string) (*AgentID, error) {
	data, err := base58.Decode(s)
	if err != nil {
		return nil, err
	}
	return AgentIDFromBytes(data)
}

// NewAgentIDFromString parses the human-readable string representation
func NewAgentIDFromString(s string) (*AgentID, error) {
	if !strings.HasPrefix(s, "A/") {
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
	if a.a == nil {
		panic("AgentID.Bytes: address == nil")
	}
	mu := marshalutil.New()
	mu.Write(a.a)
	mu.Write(a.h)
	return mu.Bytes()
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

func (a *AgentID) IsNil() bool {
	return a.Equals(&NilAgentID)
}
