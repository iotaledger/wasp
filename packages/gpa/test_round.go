// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"

	"golang.org/x/xerrors"
)

// A protocol for testing infrastructure.
// A peer outputs true when it receives a message from each peer.
type testRound struct {
	me       NodeID
	nodeIDs  []NodeID
	received map[NodeID]bool
}

var _ GPA = &testRound{}

func NewTestRound(nodeIDs []NodeID, me NodeID) GPA {
	return NewOwnHandler(me, &testRound{me: me, nodeIDs: nodeIDs, received: map[NodeID]bool{}})
}

func (tr *testRound) Input(input Input) []Message {
	msgs := make([]Message, len(tr.nodeIDs))
	for i := range msgs {
		msgs[i] = &testRoundMsg{recipient: tr.nodeIDs[i]}
	}
	return msgs
}

func (tr *testRound) Message(msg Message) []Message {
	from := msg.(*testRoundMsg).sender
	if tr.received[from] {
		panic(xerrors.Errorf("duplicate message"))
	}
	tr.received[from] = true
	return NoMessages()
}

func (tr *testRound) Output() Output {
	if len(tr.received) == len(tr.nodeIDs) {
		output := true
		return &output
	}
	return nil
}

func (tr *testRound) StatusString() string {
	return fmt.Sprintf("{testRound, received=%v}", tr.received)
}

func (tr *testRound) UnmarshalMessage(data []byte) (Message, error) {
	panic(xerrors.Errorf("not implemented"))
}

type testRoundMsg struct {
	recipient NodeID
	sender    NodeID
}

func (m *testRoundMsg) Recipient() NodeID {
	return m.recipient
}

func (m *testRoundMsg) SetSender(sender NodeID) {
	m.sender = sender
}

func (m *testRoundMsg) MarshalBinary() ([]byte, error) {
	panic(xerrors.Errorf("should be not used"))
}
