// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestAckHandler(t *testing.T) {
	t.Parallel()
	n := 10
	nodeIDs := MakeTestNodeIDs(n)
	nodesAH := map[NodeID]AckHandler{}
	nodes := map[NodeID]GPA{}
	inputs := map[NodeID]Input{}
	for _, nid := range nodeIDs {
		nodesAH[nid] = NewAckHandler(nid, NewTestRound(nodeIDs, nid), 10*time.Millisecond)
		nodes[nid] = nodesAH[nid]
		inputs[nid] = nil
	}
	tc := NewTestContext(nodes).
		WithInputs(inputs).
		WithInputProbability(0.5).
		WithMessageDeliveryProbability(0.5) // NOTE: The AckHandler has to compensate this.
	tc.RunAll()
	//
	// Tick the timer until all the messages are delivered.
	for {
		allCompleted := true
		for _, n := range nodes {
			if n.Output() == nil {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			for _, n := range nodes {
				require.True(t, *n.Output().(*bool))
			}
			break
		}
		timestamp := time.Now()
		for _, nid := range nodeIDs {
			tc.WithInput(nid, nodesAH[nid].MakeTickInput(timestamp))
		}
		tc.RunAll()
	}
}

func TestAckHandlerBatchCodec(t *testing.T) {
	v := ackHandlerBatch{
		id: lo.ToPtr(42),
		msgs: []Message{
			&TestMessage{ID: 50},
			&TestMessage{ID: 100},
		},
		acks:      []int{1, 2, 3},
		nestedGPA: &testGPA{},
	}
	vEnc := bcs.MustMarshal(&v)

	vDec := ackHandlerBatch{
		nestedGPA: &testGPA{},
	}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)

	require.Equal(t, v, vDec, vEnc)
}

type testGPA struct {
	GPA
}

var _ GPA = &testGPA{}

func (g *testGPA) UnmarshalMessage(data []byte) (Message, error) {
	return bcs.Unmarshal[TestNestedMessages](data)
}

type TestNestedMessages interface {
	Message
}

func init() {
	bcs.RegisterEnumType1[TestNestedMessages, *TestMessage]()
}
