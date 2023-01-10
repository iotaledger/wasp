// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package amDist

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

// Send by a node which has a chain enabled to a node it considers an access node.
type msgAccess struct {
	gpa.BasicMessage
	senderLClock    int
	receiverLClock  int
	accessForChains []isc.ChainID
	serverForChains []isc.ChainID
}

var _ gpa.Message = &msgAccess{}

func newMsgAccess(
	recipient gpa.NodeID,
	senderLClock, receiverLClock int,
	accessForChains []isc.ChainID,
	serverForChains []isc.ChainID,
) gpa.Message {
	return &msgAccess{
		BasicMessage:    gpa.NewBasicMessage(recipient),
		senderLClock:    senderLClock,
		receiverLClock:  receiverLClock,
		accessForChains: accessForChains,
		serverForChains: serverForChains,
	}
}

func (m *msgAccess) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if err := util.WriteByte(w, msgTypeAccess); err != nil {
		return nil, err
	}
	if err := util.WriteUint32(w, uint32(m.senderLClock)); err != nil {
		return nil, err
	}
	if err := util.WriteUint32(w, uint32(m.receiverLClock)); err != nil {
		return nil, err
	}
	if err := util.WriteUint32(w, uint32(len(m.accessForChains))); err != nil {
		return nil, err
	}
	for i := range m.accessForChains {
		if err := util.WriteBytes8(w, m.accessForChains[i].Bytes()); err != nil {
			return nil, err
		}
	}
	if err := util.WriteUint32(w, uint32(len(m.serverForChains))); err != nil {
		return nil, err
	}
	for i := range m.serverForChains {
		if err := util.WriteBytes8(w, m.serverForChains[i].Bytes()); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (m *msgAccess) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	var u32 uint32
	if msgType, err := util.ReadByte(r); err != nil || msgType != msgTypeAccess {
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected message type: %v", msgType)
	}
	//
	// senderLClock
	if err := util.ReadUint32(r, &u32); err != nil {
		return err
	}
	m.senderLClock = int(u32)
	//
	// receiverLClock
	if err := util.ReadUint32(r, &u32); err != nil {
		return err
	}
	m.receiverLClock = int(u32)
	//
	// accessForChains
	if err := util.ReadUint32(r, &u32); err != nil {
		return err
	}
	m.accessForChains = make([]isc.ChainID, u32)
	for i := range m.accessForChains {
		val, err := util.ReadBytes8(r)
		if err != nil {
			return err
		}
		chainID, err := isc.ChainIDFromBytes(val)
		if err != nil {
			return err
		}
		m.accessForChains[i] = chainID
	}
	//
	// serverForChains
	if err := util.ReadUint32(r, &u32); err != nil {
		return err
	}
	m.serverForChains = make([]isc.ChainID, u32)
	for i := range m.serverForChains {
		val, err := util.ReadBytes8(r)
		if err != nil {
			return err
		}
		chainID, err := isc.ChainIDFromBytes(val)
		if err != nil {
			return err
		}
		m.serverForChains[i] = chainID
	}
	return nil
}
