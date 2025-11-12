// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"
	"math/rand"
	"sort"

	"github.com/samber/lo"
)

type TestContextNewFunctors[Obj any, Input any, MsgPayload any] struct {
	ApplyInput       func(obj Obj, input Input) []PayloadOut[MsgPayload]
	ApplyMessage     func(obj Obj, sender NodeID, msg MsgPayload) []PayloadOut[MsgPayload]
	Output           func(obj Obj) any
	StatusString     func(obj Obj) string
	MarshalPayload   func(msg MsgPayload) ([]byte, error)
	UnmarshalPayload func(obj Obj, data []byte) (MsgPayload, error)
}

type pendingMessageNew[MsgPayload any] struct {
	Recipient NodeID
	Msg       PayloadIn[MsgPayload]
}

// TestContextNew imitates a cluster of nodes and the medium performing the message exchange.
// Inputs are processes in-order for each node individually.
type TestContextNew[Obj any, Input any, MsgPayload any] struct {
	functors        TestContextNewFunctors[Obj, Input, MsgPayload]
	nodes           map[NodeID]Obj                     // Nodes to test.
	inputs          map[NodeID][]Input                 // Not yet provided inputs.
	inputCh         <-chan map[NodeID]Input            // A way to provide additional inputs w/o synchronizing other parts.
	inputProb       float64                            // A probability to process input, instead of a message (if any).
	inputCount      int                                // Number if inputs still not delivered.
	outputHandler   func(nodeID NodeID, output Output) // User can check outputs w/o synchronizing other parts.
	msgDeliveryProb float64                            // A probability to deliver a message (to not discard/loose it).
	msgSerialize    bool                               // Use serialization/deserialization when delivering the messages?
	msgs            []pendingMessageNew[MsgPayload]    // Not yet delivered messages.
	msgsSent        int                                // Stats.
	msgsRecv        int                                // Stats.
	bytesRecv       int
}

func NewTestContextNew[Obj any, Input any, MsgPayload any](nodes map[NodeID]Obj, functors TestContextNewFunctors[Obj, Input, MsgPayload]) *TestContextNew[Obj, Input, MsgPayload] {
	inputs := map[NodeID][]Input{}
	for n := range nodes {
		inputs[n] = []Input{}
	}
	tc := TestContextNew[Obj, Input, MsgPayload]{
		functors:        functors,
		msgSerialize:    true,
		nodes:           nodes,
		inputs:          inputs,
		inputProb:       1.0,
		inputCount:      0,
		msgDeliveryProb: 1.0,
		msgs:            []pendingMessageNew[MsgPayload]{},
	}
	return &tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithoutSerialization() *TestContextNew[Obj, Input, MsgPayload] {
	tc.msgSerialize = false
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) MsgCounts() (int, int) {
	return tc.msgsSent, tc.msgsRecv
}

// AddInputs adds new inputs to the existing set.
// The inputs will be overridden, if exist for the same nodes.
func (tc *TestContextNew[Obj, Input, MsgPayload]) AddInputs(inputs map[NodeID]Input) {
	for nid := range inputs {
		tc.inputs[nid] = append(tc.inputs[nid], inputs[nid])
	}
	tc.inputCount += len(inputs)
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithInput(nodeID NodeID, input Input) *TestContextNew[Obj, Input, MsgPayload] {
	tc.AddInputs(map[NodeID]Input{nodeID: input})
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithInputs(inputs map[NodeID]Input) *TestContextNew[Obj, Input, MsgPayload] {
	tc.AddInputs(inputs)
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithInputChannel(inputCh <-chan map[NodeID]Input) *TestContextNew[Obj, Input, MsgPayload] {
	tc.inputCh = inputCh
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithInputProbability(inputProb float64) *TestContextNew[Obj, Input, MsgPayload] {
	tc.inputProb = inputProb
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithMessageDeliveryProbability(msgDeliveryProb float64) *TestContextNew[Obj, Input, MsgPayload] {
	tc.msgDeliveryProb = msgDeliveryProb
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithMessages(recipient NodeID, msgs []PayloadIn[MsgPayload]) *TestContextNew[Obj, Input, MsgPayload] {
	tc.addMessages(lo.Map(msgs, func(m PayloadIn[MsgPayload], _ int) pendingMessageNew[MsgPayload] {
		return pendingMessageNew[MsgPayload]{Recipient: recipient, Msg: m}
	}))
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) addMessages(msgs []pendingMessageNew[MsgPayload]) {
	tc.msgsSent += len(msgs)
	tc.msgs = append(tc.msgs, msgs...)
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithMessage(recipient NodeID, msg PayloadIn[MsgPayload]) *TestContextNew[Obj, Input, MsgPayload] {
	tc.msgsSent++
	tc.msgs = append(tc.msgs, pendingMessageNew[MsgPayload]{Recipient: recipient, Msg: msg})
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) WithOutputHandler(outputHandler func(nodeID NodeID, output Output)) *TestContextNew[Obj, Input, MsgPayload] {
	tc.outputHandler = outputHandler
	return tc
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) RunUntil(predicate func() bool) {
	loop := make(chan bool, 1)
	loop <- true
	keepLooping := func() {
		if len(loop) == 0 {
			loop <- true
		}
	}
	for {
		select {
		case inputs, ok := <-tc.inputCh:
			keepLooping()
			if !ok {
				tc.inputCh = nil
				continue
			}
			if len(inputs) == 0 {
				continue
			}
			for nid, input := range inputs {
				tc.inputs[nid] = append(tc.inputs[nid], input)
			}
			tc.inputCount += len(inputs)
		case <-loop:
			if predicate() {
				return
			}
			tc.tryProcessInput()   // Try provide an input, if any and we are lucky.
			tc.tryProcessMessage() // Otherwise just process the messages.
			if len(tc.msgs) > 0 || tc.inputCount > 0 {
				// We can proceed with looping.
				loop <- true
				continue
			}
			if tc.inputCh == nil {
				// Channels are closed and there is no more inputs or messages. Stop it.
				return
			}
			// Otherwise we have to wait for something from channels.
		}
	}
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) tryProcessInput() {
	if tc.inputCount > 0 && (rand.Float64() <= tc.inputProb || len(tc.msgs) == 0) {
		rnd := rand.Intn(tc.inputCount)
		var rndNID NodeID
		var rndInp Input
		for nodeID, nodeInputs := range tc.inputs {
			if rnd >= len(nodeInputs) {
				rnd -= len(nodeInputs)
				continue
			}
			rndNID = nodeID
			rndInp = nodeInputs[0]
			tc.inputs[nodeID] = nodeInputs[1:] // Take them in order.
			break
		}
		tc.inputCount--

		// fmt.Printf("-> %s :: INPUT %s\n", rndNID.ShortString(), rndInp)
		msgs := tc.functors.ApplyInput(tc.nodes[rndNID], rndInp)
		tc.addMessages(lo.Map(msgs, func(m PayloadOut[MsgPayload], _ int) pendingMessageNew[MsgPayload] {
			return pendingMessageNew[MsgPayload]{Recipient: m.Recipient, Msg: NewPayloadIn(rndNID, m.Payload)}
		}))
		tc.tryCallOutputHandler(rndNID)
	}
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) tryProcessMessage() {
	if len(tc.msgs) == 0 {
		return
	}

	// select a random message, swap it with the last one and decrease the slice length
	rnd := rand.Intn(len(tc.msgs))
	pendingMsg := tc.msgs[rnd]
	tc.msgs[rnd] = tc.msgs[len(tc.msgs)-1]
	tc.msgs = tc.msgs[:len(tc.msgs)-1]

	tc.msgsRecv++
	if rand.Float64() > tc.msgDeliveryProb {
		// message dropped
		return
	}

	nid := pendingMsg.Recipient
	msg := pendingMsg.Msg
	if tc.msgSerialize {
		msgBytes := lo.Must(tc.functors.MarshalPayload(msg.Payload))
		tc.bytesRecv += len(msgBytes)
		m, err := tc.functors.UnmarshalPayload(tc.nodes[nid], msgBytes)
		if err != nil {
			// E.g. silent node cannot decode messages.
			return
		}
		msg = NewPayloadIn(msg.Sender, m)
	}
	// fmt.Printf("%s -> %s :: %s (count: %d / %d bytes)\n", msg.Sender.ShortString(), nid.ShortString(), msg.Payload, tc.msgsRecv, tc.bytesRecv)
	msgs := tc.functors.ApplyMessage(tc.nodes[nid], msg.Sender, msg.Payload)
	tc.addMessages(lo.Map(msgs, func(m PayloadOut[MsgPayload], _ int) pendingMessageNew[MsgPayload] {
		return pendingMessageNew[MsgPayload]{Recipient: m.Recipient, Msg: NewPayloadIn(nid, m.Payload)}
	}))
	tc.tryCallOutputHandler(nid)
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) tryCallOutputHandler(nid NodeID) {
	out := tc.functors.Output(tc.nodes[nid])
	if out != nil && tc.outputHandler != nil {
		tc.outputHandler(nid, out)
	}
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) RunAll() {
	tc.RunUntil(tc.OutOfMessagesPredicate())
}

// NumberOfOutputs returns a number of non-nil outputs.
func (tc *TestContextNew[Obj, Input, MsgPayload]) NumberOfOutputs() int {
	outNum := 0
	for _, node := range tc.nodes {
		output := tc.functors.Output(node)
		if output != nil {
			outNum++
		}
	}
	return outNum
}

// NumberOfOutputsPredicate runs until there will be at least outNum of non-nil outputs generated.
func (tc *TestContextNew[Obj, Input, MsgPayload]) NumberOfOutputsPredicate(outNum int) func() bool {
	return func() bool {
		return tc.NumberOfOutputs() >= outNum
	}
}

// OutOfMessagesPredicate runs until all the messages will be processed.
func (tc *TestContextNew[Obj, Input, MsgPayload]) OutOfMessagesPredicate() func() bool {
	return func() bool { return false }
}

func (tc *TestContextNew[Obj, Input, MsgPayload]) PrintAllStatusStrings(prefix string, logFunc func(format string, args ...any)) {
	logFunc("TC[%p] Status, |inputs|=%v, inputsCh=%v, |msgs|=%v", tc, tc.inputCount, tc.inputCh != nil, len(tc.msgs))
	keys := []NodeID{}
	for nid := range tc.nodes {
		keys = append(keys, nid)
	}
	// Print them sorted.
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i][:], keys[j][:]) < 0
	})
	for _, nidStr := range keys {
		logFunc("TC[%p] %v [node=%v]: %v", tc, prefix, nidStr, tc.functors.StatusString(tc.nodes[nidStr]))
	}
}

func ToAnyPayloadsOut[Payload any](payloads []PayloadOut[Payload]) []PayloadOut[any] {
	res := make([]PayloadOut[any], len(payloads))
	for i, p := range payloads {
		res[i] = PayloadOut[any]{
			Recipient: p.Recipient,
			Payload:   p.Payload,
		}
	}
	return res
}
