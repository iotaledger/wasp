// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"math/rand"
)

// Imitates a cluster of nodes and the medium performing the message exchange.
type TestContext struct {
	nodes  map[NodeID]GPA
	inputs map[NodeID]Input
	msgs   []Message
}

func NewTestContext(nodes map[NodeID]GPA) *TestContext {
	tc := TestContext{
		nodes:  nodes,
		inputs: map[NodeID]Input{},
		msgs:   []Message{},
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

func (tc *TestContext) SendMessages(msgs []Message) {
	tc.msgs = append(tc.msgs, msgs...)
}

func (tc *TestContext) RunUntil(inputProb float64, predicate func() bool) {
	for (len(tc.msgs) > 0 || len(tc.inputs) > 0) && !predicate() {
		//
		// Try provide an input, if any and we are lucky.
		inputRand := rand.Float64()
		if len(tc.inputs) > 0 && (inputRand <= inputProb || len(tc.msgs) == 0) {
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
		tc.msgs = append(tc.msgs, tc.setMessageSender(nid, tc.nodes[nid].Message(msg))...)
	}
}

func (tc *TestContext) RunAll(inputProb ...float32) {
	if len(inputProb) == 0 {
		tc.RunUntil(1.0, tc.OutOfMessagesPredicate())
	}
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
