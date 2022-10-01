// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"encoding"

	"github.com/iotaledger/wasp/packages/gpa"
)

// type BBAMsg interface {
// 	gpa.Message
// 	SetRecipient(gpa.NodeID)
// }

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
	for _, n := range recipients {
		msgs.Add(&msgVote{recipient: n, round: round, voteType: voteType, value: value})
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
	panic("to be implemented") // TODO: Impl MarshalBinary
}

func (m *msgVote) UnmarshalBinary(data []byte) error {
	panic("to be implemented") // TODO: Impl UnmarshalBinary
}
