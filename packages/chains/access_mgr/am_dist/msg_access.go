// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package am_dist

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
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
	if err := rwutil.WriteByte(w, msgTypeAccess); err != nil {
		return nil, err
	}
	if err := rwutil.WriteUint32(w, uint32(m.senderLClock)); err != nil {
		return nil, err
	}
	if err := rwutil.WriteUint32(w, uint32(m.receiverLClock)); err != nil {
		return nil, err
	}
	if err := rwutil.WriteUint32(w, uint32(len(m.accessForChains))); err != nil {
		return nil, err
	}
	for i := range m.accessForChains {
		if err := rwutil.WriteBytes(w, m.accessForChains[i].Bytes()); err != nil {
			return nil, err
		}
	}
	if err := rwutil.WriteUint32(w, uint32(len(m.serverForChains))); err != nil {
		return nil, err
	}
	for i := range m.serverForChains {
		if err := rwutil.WriteBytes(w, m.serverForChains[i].Bytes()); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

//nolint:govet
func (m *msgAccess) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := rwutil.ReadByte(r)
	if err != nil || msgType != msgTypeAccess {
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected message type: %v", msgType)
	}
	//
	// senderLClock
	var u32 uint32
	if u32, err = rwutil.ReadUint32(r); err != nil {
		return err
	}
	m.senderLClock = int(u32)
	//
	// receiverLClock
	if u32, err = rwutil.ReadUint32(r); err != nil {
		return err
	}
	m.receiverLClock = int(u32)
	//
	// accessForChains
	if u32, err = rwutil.ReadUint32(r); err != nil {
		return err
	}
	m.accessForChains = make([]isc.ChainID, u32)
	for i := range m.accessForChains {
		val, err := rwutil.ReadBytes(r)
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
	if u32, err = rwutil.ReadUint32(r); err != nil {
		return err
	}
	m.serverForChains = make([]isc.ChainID, u32)
	for i := range m.serverForChains {
		val, err := rwutil.ReadBytes(r)
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
