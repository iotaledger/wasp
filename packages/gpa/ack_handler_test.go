// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
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
	testMsgs := []ackHandlerBatch{
		{
			id: lo.ToPtr(42),
			msgs: []Message{
				&TestMessage{ID: 50},
				&TestMessage{ID: 100},
			},
			acks:      []int{1, 2, 3},
			nestedGPA: &testGPA{},
		},
		{
			id:        lo.ToPtr(42),
			msgs:      []Message{},
			acks:      []int{1, 2, 3},
			nestedGPA: &testGPA{},
		},
		{
			id:        lo.ToPtr(42),
			msgs:      nil,
			acks:      []int{1, 2, 3},
			nestedGPA: &testGPA{},
		},
	}

	for _, v := range testMsgs {
		vEnc := bcs.MustMarshal(&v)
		vDec := bcs.MustUnmarshalInto(vEnc, &ackHandlerBatch{nestedGPA: &testGPA{}})
		if len(v.msgs) == 0 {
			require.Len(t, vDec.msgs, 0)
			vDec.msgs = v.msgs
		}
		require.Equal(t, v, *vDec, vEnc)

		v.id = nil
		vEnc = bcs.MustMarshal(&v)
		vDec = bcs.MustUnmarshalInto(vEnc, &ackHandlerBatch{nestedGPA: &testGPA{}})
		if len(v.msgs) == 0 {
			require.Len(t, vDec.msgs, 0)
			vDec.msgs = v.msgs
		}
		require.Equal(t, v, *vDec, vEnc)
	}
}

type testGPA struct {
	GPA
}

var _ GPA = &testGPA{}

func (g *testGPA) UnmarshalMessage(data []byte) (Message, error) {
	return UnmarshalMessage(data, Mapper{
		msgTypeTest: func() Message { return &TestMessage{} },
	}, nil)
}

func TestAckHandlerResetCodec(t *testing.T) {
	bcs.TestCodecAndHash(t, ackHandlerReset{
		response: true,
		latestID: 123,
	}, "85add3e79841")
}
