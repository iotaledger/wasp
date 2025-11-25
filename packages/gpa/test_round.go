// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"errors"
	"fmt"
)

const (
	msgTypeTestRound MessageType = iota
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

func (tr *testRound) Input(input Input) []MessageOut {
	msgs := make([]MessageOut, len(tr.nodeIDs))
	for i := range msgs {
		msgs[i] = NewMessageOut(tr.nodeIDs[i], &testRoundMsg{})
	}
	return msgs
}

func (tr *testRound) Message(msg MessageIn[any]) []MessageOut {
	from := msg.Sender
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

func (tr *testRound) MarshalPayload(payload any) ([]byte, error) {
	switch p := payload.(type) {
	case *testRoundMsg:
		return MarshalPayload(msgTypeTestRound, p)
	default:
		panic(fmt.Errorf("testRound: unknown payload type %T", payload))
	}
}

func (tr *testRound) UnmarshalPayload(data []byte) (any, error) {
	return UnmarshalPayload(data, PayloadAllocator{
		msgTypeTestRound: func() any { return &testRoundMsg{} },
	})
}

type testRoundMsg struct{}
