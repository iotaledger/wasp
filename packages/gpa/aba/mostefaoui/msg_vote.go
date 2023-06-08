// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"bytes"
	"encoding"
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgVoteType byte

const (
	BVAL msgVoteType = iota
	AUX
)

type msgVote struct {
	sender    gpa.NodeID
	recipient gpa.NodeID
	round     int
	voteType  msgVoteType
	value     bool
}

var (
	_ gpa.Message                = &msgVote{}
	_ encoding.BinaryUnmarshaler = &msgVote{}
)

func multicastMsgVote(recipients []gpa.NodeID, round int, voteType msgVoteType, value bool) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, nid := range recipients {
		msgs.Add(&msgVote{recipient: nid, round: round, voteType: voteType, value: value})
	}
	return msgs
}

func (m *msgVote) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgVote) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgVote) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if err := rwutil.WriteByte(w, msgTypeVote); err != nil {
		return nil, err
	}
	if err := rwutil.WriteUint16(w, uint16(m.round)); err != nil {
		return nil, err
	}
	if err := rwutil.WriteByte(w, byte(m.voteType)); err != nil {
		return nil, err
	}
	if err := rwutil.WriteBool(w, m.value); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgVote) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := rwutil.ReadByte(r)
	if err != nil {
		return err
	}
	if msgType != msgTypeVote {
		return fmt.Errorf("expected msgTypeVote, got %v", msgType)
	}
	var round uint16
	if err2 := rwutil.ReadUint16ByRef(r, &round); err2 != nil {
		return err2
	}
	m.round = int(round)
	voteType, err := rwutil.ReadByte(r)
	if err != nil {
		return err
	}
	m.voteType = msgVoteType(voteType)
	if err := rwutil.ReadBoolByRef(r, &m.value); err != nil {
		return err
	}
	return nil
}
