// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"strings"

	"github.com/iotaledger/wasp/packages/iscp/placeholders"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// AgentID represents address on the ledger with optional hname
// If address is and alias address and hname is != 0 the agent id is interpreted as
// ad contract id
type AgentID struct {
	a iotago.Address
	h Hname
}

var NilAgentID AgentID

func init() {
	NilAgentID = AgentID{
		a: nil,
		h: 0,
	}
}

// NewAgentID makes new AgentID
func NewAgentID(addr iotago.Address, hname Hname) *AgentID {
	return &AgentID{
		a: addr,
		h: hname,
	}
}

func AgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (*AgentID, error) {
	var err error
	ret := &AgentID{}
	// TODO
	//if ret.a, err = ledgerstate.AddressFromMarshalUtil(mu); err != nil {
	//	return nil, err
	//}
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
	// TODO placeholder
	addr, err := placeholders.AddressFromStringTmp(parts[0]) // old ledgerstate.AddressFromBase58EncodedString(parts[0])
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
	raddr := RandomChainID()
	return NewAgentID(raddr.AsAddress(), Hn("testName"))
}

func (a *AgentID) Address() iotago.Address {
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
	// TODO placeholder
	placeholders.WriteAddressToMarshalUtil(mu, a.a)
	mu.Write(a.h)
	return mu.Bytes()
}

func (a *AgentID) Equals(a1 *AgentID) bool {
	if !a.a.Equal(a1.a) {
		return false
	}
	if a.h != a1.h {
		return false
	}
	return true
}

// String human readable string
func (a *AgentID) String() string {
	return "A/" + a.a.String() + "::" + a.h.String()
}

func (a *AgentID) Base58() string {
	return base58.Encode(a.Bytes())
}

func (a *AgentID) IsNil() bool {
	return a.Equals(&NilAgentID)
}
