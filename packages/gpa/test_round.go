// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"errors"
	"fmt"
	"io"
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

func (tr *testRound) Input(input Input) OutMessages {
	msgs := make([]Message, len(tr.nodeIDs))
	for i := range msgs {
		msgs[i] = &testRoundMsg{BasicMessage: NewBasicMessage(tr.nodeIDs[i])}
	}
	return NoMessages().AddMany(msgs)
}

func (tr *testRound) Message(msg Message) OutMessages {
	from := msg.(*testRoundMsg).sender
	if tr.received[from] {
		panic(errors.New("duplicate message"))
	}
	tr.received[from] = true
	return nil
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
	panic(errors.New("not implemented"))
}

type testRoundMsg struct {
	BasicMessage
}

var _ Message = new(testRoundMsg)

func (msg *testRoundMsg) Read(r io.Reader) error {
	panic(errors.New("should be not used"))
}

func (msg *testRoundMsg) Write(w io.Writer) error {
	panic(errors.New("should be not used"))
}
