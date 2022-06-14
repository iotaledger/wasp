// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"math/rand"
)

// Imitates a cluster of nodes and the medium performing the message exchange.
type TestContext struct {
	nodes           map[NodeID]GPA   // Nodes to test.
	inputs          map[NodeID]Input // Not yet provided inputs.
	inputProb       float64          // A probability to process input, instead of a message (if any).
	msgDeliveryProb float64          // A probability to deliver a message (to not discard/loose it).
	msgs            []Message        // Not yet delivered messages.
}

func NewTestContext(nodes map[NodeID]GPA) *TestContext {
	tc := TestContext{
		nodes:           nodes,
		inputs:          map[NodeID]Input{},
		inputProb:       1.0,
		msgDeliveryProb: 1.0,
		msgs:            []Message{},
	}
	return &tc
}

// Will add new inputs to the existing set.
// The inputs will be overridden, if exist for the same nodes.
func (tc *TestContext) AddInputs(inputs map[NodeID]Input) {
	for nid := range inputs {
		tc.inputs[nid] = inputs[nid]
	}
}

func (tc *TestContext) WithInputs(inputs map[NodeID]Input) *TestContext {
	tc.AddInputs(inputs)
	return tc
}

func (tc *TestContext) WithInputProbability(inputProb float64) *TestContext {
	tc.inputProb = inputProb
	return tc
}

func (tc *TestContext) WithMessageDeliveryProbability(msgDeliveryProb float64) *TestContext {
	tc.msgDeliveryProb = msgDeliveryProb
	return tc
}

func (tc *TestContext) WithMessages(msgs []Message) *TestContext {
	tc.msgs = append(tc.msgs, msgs...)
	return tc
}

func (tc *TestContext) WithCall(call func() []Message) *TestContext {
	msgs := call()
	return tc.WithMessages(msgs)
}

func (tc *TestContext) RunUntil(predicate func() bool) {
	for (len(tc.msgs) > 0 || len(tc.inputs) > 0) && !predicate() {
		//
		// Try provide an input, if any and we are lucky.
		if len(tc.inputs) > 0 && (rand.Float64() <= tc.inputProb || len(tc.msgs) == 0) {
			nids := []NodeID{}
			for nid := range tc.inputs {
				nids = append(nids, nid)
			}
			nid := nids[rand.Intn(len(nids))]
			tc.msgs = append(tc.msgs, tc.setMessageSender(nid, tc.nodes[nid].Input(tc.inputs[nid]))...)
			delete(tc.inputs, nid)
		}
		//
		// Otherwise just process the messages.
		msgIdx := rand.Intn(len(tc.msgs))
		msg := tc.msgs[msgIdx]
		nid := msg.Recipient()
		tc.msgs = append(tc.msgs[:msgIdx], tc.msgs[msgIdx+1:]...)
		if rand.Float64() <= tc.msgDeliveryProb { // Deliver some messages.
			tc.msgs = append(tc.msgs, tc.setMessageSender(nid, tc.nodes[nid].Message(msg))...)
		}
	}
}

func (tc *TestContext) RunAll() {
	tc.RunUntil(tc.OutOfMessagesPredicate())
}

// Returns a number of non-nil outputs.
func (tc *TestContext) NumberOfOutputs() int {
	outNum := 0
	for _, node := range tc.nodes {
		if node.Output() != nil {
			outNum++
		}
	}
	return outNum
}

// Will run until there will be at least outNum of non-nil outputs generated.
func (tc *TestContext) NumberOfOutputsPredicate(outNum int) func() bool {
	return func() bool {
		return tc.NumberOfOutputs() >= outNum
	}
}

// Will run until all the messages will be processed.
func (tc *TestContext) OutOfMessagesPredicate() func() bool {
	return func() bool { return false }
}

func (tc *TestContext) setMessageSender(sender NodeID, msgs []Message) []Message {
	for i := range msgs {
		msgs[i].SetSender(sender)
	}
	return msgs
}
