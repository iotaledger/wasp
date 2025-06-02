// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"
	"math/rand"
	"sort"

	"github.com/samber/lo"
)

// TestContext imitates a cluster of nodes and the medium performing the message exchange.
// Inputs are processes in-order for each node individually.
type TestContext struct {
	nodes           map[NodeID]GPA                     // Nodes to test.
	inputs          map[NodeID][]Input                 // Not yet provided inputs.
	inputCh         <-chan map[NodeID]Input            // A way to provide additional inputs w/o synchronizing other parts.
	inputProb       float64                            // A probability to process input, instead of a message (if any).
	inputCount      int                                // Number if inputs still not delivered.
	outputHandler   func(nodeID NodeID, output Output) // User can check outputs w/o synchronizing other parts.
	msgDeliveryProb float64                            // A probability to deliver a message (to not discard/loose it).
	msgSerialize    bool                               // Use serialization/deserialization when delivering the messages?
	msgCh           <-chan SenderMessage               // A way to provide additional messages w/o synchronizing other parts.
	msgs            []SenderMessage                    // Not yet delivered messages.
	msgsSent        int                                // Stats.
	msgsRecv        int                                // Stats.
}

func NewTestContext(nodes map[NodeID]GPA) *TestContext {
	inputs := map[NodeID][]Input{}
	for n := range nodes {
		inputs[n] = []Input{}
	}
	tc := TestContext{
		msgSerialize:    true,
		nodes:           nodes,
		inputs:          inputs,
		inputProb:       1.0,
		inputCount:      0,
		msgDeliveryProb: 1.0,
		msgs:            []SenderMessage{},
	}
	return &tc
}

func (tc *TestContext) WithoutSerialization() *TestContext {
	tc.msgSerialize = false
	return tc
}

func (tc *TestContext) MsgCounts() (int, int) {
	return tc.msgsSent, tc.msgsRecv
}

// AddInputs adds new inputs to the existing set.
// The inputs will be overridden, if exist for the same nodes.
func (tc *TestContext) AddInputs(inputs map[NodeID]Input) {
	for nid := range inputs {
		tc.inputs[nid] = append(tc.inputs[nid], inputs[nid])
	}
	tc.inputCount += len(inputs)
}

func (tc *TestContext) WithInput(nodeID NodeID, input Input) *TestContext {
	tc.AddInputs(map[NodeID]Input{nodeID: input})
	return tc
}

func (tc *TestContext) WithInputs(inputs map[NodeID]Input) *TestContext {
	tc.AddInputs(inputs)
	return tc
}

func (tc *TestContext) WithInputChannel(inputCh <-chan map[NodeID]Input) *TestContext {
	tc.inputCh = inputCh
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

func (tc *TestContext) WithMessages(sender NodeID, msgs []Message) *TestContext {
	tc.msgsSent += len(msgs)
	tc.msgs = append(tc.msgs, tc.setMessageSender(sender, NoMessages().AddMany(msgs))...)
	return tc
}

func (tc *TestContext) WithMessage(sender NodeID, msg Message) *TestContext {
	tc.msgsSent++
	tc.msgs = append(tc.msgs, tc.setMessageSender(sender, NoMessages().Add(msg))...)
	return tc
}

func (tc *TestContext) WithMessageChannel(msgCh <-chan SenderMessage) *TestContext {
	tc.msgCh = msgCh
	return tc
}

func (tc *TestContext) WithOutputHandler(outputHandler func(nodeID NodeID, output Output)) *TestContext {
	tc.outputHandler = outputHandler
	return tc
}

func (tc *TestContext) WithCall(sender NodeID, call func() []Message) *TestContext {
	msgs := call()
	return tc.WithMessages(sender, msgs)
}

func (tc *TestContext) RunUntil(predicate func() bool) {
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
		case msg, ok := <-tc.msgCh:
			keepLooping()
			if !ok {
				tc.msgCh = nil
				continue
			}
			tc.msgs = append(tc.msgs, msg)
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
			if tc.inputCh == nil && tc.msgCh == nil {
				// Channels are closed and there is no more inputs or messages. Stop it.
				return
			}
			// Otherwise we have to wait for something from channels.
		}
	}
}

func (tc *TestContext) tryProcessInput() {
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

		newMsgs := tc.setMessageSender(rndNID, tc.nodes[rndNID].Input(rndInp))
		if newMsgs != nil {
			tc.msgsSent += len(newMsgs)
			tc.msgs = append(tc.msgs, newMsgs...)
		}
		tc.tryCallOutputHandler(rndNID)
	}
}

func (tc *TestContext) tryProcessMessage() {
	if len(tc.msgs) == 0 {
		return
	}
	msgIdx := rand.Intn(len(tc.msgs))
	msg := tc.msgs[msgIdx]
	nid := msg.Message.Recipient()
	tc.msgs = append(tc.msgs[:msgIdx], tc.msgs[msgIdx+1:]...)
	tc.msgsRecv++
	if rand.Float64() <= tc.msgDeliveryProb { // Deliver some messages.
		gpaMsg := msg.Message
		if tc.msgSerialize {
			msgBytes := lo.Must(MarshalMessage(msg.Message))
			if m, err := tc.nodes[nid].UnmarshalMessage(msgBytes); err == nil {
				gpaMsg = m
				gpaMsg.SetSender(msg.Sender)
			} else {
				// E.g. silent node cannot decode messages.
				gpaMsg = nil
			}
		}
		if gpaMsg != nil {
			newMsgs := tc.setMessageSender(nid, tc.nodes[nid].Message(gpaMsg))
			if newMsgs != nil {
				tc.msgsSent += len(newMsgs)
				tc.msgs = append(tc.msgs, newMsgs...)
			}
			tc.tryCallOutputHandler(nid)
		}
	}
}

func (tc *TestContext) tryCallOutputHandler(nid NodeID) {
	out := tc.nodes[nid].Output()
	if out != nil && tc.outputHandler != nil {
		tc.outputHandler(nid, out)
	}
}

func (tc *TestContext) RunAll() {
	tc.RunUntil(tc.OutOfMessagesPredicate())
}

// NumberOfOutputs returns a number of non-nil outputs.
func (tc *TestContext) NumberOfOutputs() int {
	outNum := 0
	for _, node := range tc.nodes {
		if node.Output() != nil {
			outNum++
		}
	}
	return outNum
}

// NumberOfOutputsPredicate runs until there will be at least outNum of non-nil outputs generated.
func (tc *TestContext) NumberOfOutputsPredicate(outNum int) func() bool {
	return func() bool {
		return tc.NumberOfOutputs() >= outNum
	}
}

// OutOfMessagesPredicate runs until all the messages will be processed.
func (tc *TestContext) OutOfMessagesPredicate() func() bool {
	return func() bool { return false }
}

func (tc *TestContext) setMessageSender(sender NodeID, msgs OutMessages) []SenderMessage {
	if msgs == nil {
		return nil
	}
	msgArray := msgs.AsArray()
	result := make([]SenderMessage, len(msgArray))
	for i := range msgArray {
		msgArray[i].SetSender(sender)
		result[i] = SenderMessage{Sender: sender, Message: msgArray[i]}
	}
	return result
}

func (tc *TestContext) PrintAllStatusStrings(prefix string, logFunc func(format string, args ...any)) {
	logFunc("TC[%p] Status, |inputs|=%v, inputsCh=%v, |msgs|=%v, msgsCh=%v", tc, tc.inputCount, tc.inputCh != nil, len(tc.msgs), tc.msgCh != nil)
	keys := []NodeID{}
	for nid := range tc.nodes {
		keys = append(keys, nid)
	}
	// Print them sorted.
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i][:], keys[j][:]) < 0
	})
	for _, nidStr := range keys {
		logFunc("TC[%p] %v [node=%v]: %v", tc, prefix, nidStr, tc.nodes[nidStr].StatusString())
	}
}

type SenderMessage struct {
	Sender  NodeID
	Message Message
}
