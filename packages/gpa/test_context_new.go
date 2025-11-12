// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

// import (
// 	"bytes"
// 	"math/rand"
// 	"sort"

// 	"github.com/samber/lo"
// )

// type TestContextNewFunctors[Obj any, Input any, Msg any] struct {
// 	ApplyInput       func(obj Obj, input Input) []Msg
// 	ApplyMessage     func(obj Obj, msg Msg) []Msg
// 	Output           func(obj Obj) any
// 	StatusString     func(obj Obj) string
// 	MarshalMessage   func(msg Msg) ([]byte, error)
// 	UnmarshalMessage func(obj Obj, data []byte) (Msg, error)
// }

// // TestContextNew imitates a cluster of nodes and the medium performing the message exchange.
// // Inputs are processes in-order for each node individually.
// type TestContextNew[Obj any, Input any, Msg any] struct {
// 	functors        TestContextNewFunctors[Obj, Input, Msg]
// 	nodes           map[NodeID]Obj                     // Nodes to test.
// 	inputs          map[NodeID][]Input                 // Not yet provided inputs.
// 	inputCh         <-chan map[NodeID]Input            // A way to provide additional inputs w/o synchronizing other parts.
// 	inputProb       float64                            // A probability to process input, instead of a message (if any).
// 	inputCount      int                                // Number if inputs still not delivered.
// 	outputHandler   func(nodeID NodeID, output Output) // User can check outputs w/o synchronizing other parts.
// 	msgDeliveryProb float64                            // A probability to deliver a message (to not discard/loose it).
// 	msgSerialize    bool                               // Use serialization/deserialization when delivering the messages?
// 	msgCh           <-chan Senderany[Msg]              // A way to provide additional messages w/o synchronizing other parts.
// 	msgs            []Senderany[Msg]                   // Not yet delivered messages.
// 	msgsSent        int                                // Stats.
// 	msgsRecv        int                                // Stats.
// }

// func NewTestContextNew[Obj any, Input any, Msg any](
// 	nodes map[NodeID]Obj,
// 	functors TestContextNewFunctors[Obj, Input, Msg],
// ) *TestContextNew[Obj, Input, Msg] {
// 	inputs := map[NodeID][]Input{}
// 	for n := range nodes {
// 		inputs[n] = []Input{}
// 	}
// 	tc := TestContextNew[Obj, Input, Msg]{
// 		functors:        functors,
// 		msgSerialize:    true,
// 		nodes:           nodes,
// 		inputs:          inputs,
// 		inputProb:       1.0,
// 		inputCount:      0,
// 		msgDeliveryProb: 1.0,
// 		msgs:            []Senderany[Msg]{},
// 	}
// 	return &tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithoutSerialization() *TestContextNew[Obj, Input, Message] {
// 	tc.msgSerialize = false
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) MsgCounts() (int, int) {
// 	return tc.msgsSent, tc.msgsRecv
// }

// // AddInputs adds new inputs to the existing set.
// // The inputs will be overridden, if exist for the same nodes.
// func (tc *TestContextNew[Obj, Input, Message]) AddInputs(inputs map[NodeID]Input) {
// 	for nid := range inputs {
// 		tc.inputs[nid] = append(tc.inputs[nid], inputs[nid])
// 	}
// 	tc.inputCount += len(inputs)
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithInput(nodeID NodeID, input Input) *TestContextNew[Obj, Input, Message] {
// 	tc.AddInputs(map[NodeID]Input{nodeID: input})
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithInputs(inputs map[NodeID]Input) *TestContextNew[Obj, Input, Message] {
// 	tc.AddInputs(inputs)
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithInputChannel(inputCh <-chan map[NodeID]Input) *TestContextNew[Obj, Input, Message] {
// 	tc.inputCh = inputCh
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithInputProbability(inputProb float64) *TestContextNew[Obj, Input, Message] {
// 	tc.inputProb = inputProb
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithMessageDeliveryProbability(msgDeliveryProb float64) *TestContextNew[Obj, Input, Message] {
// 	tc.msgDeliveryProb = msgDeliveryProb
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithMessages(sender NodeID, msgs []Message) *TestContextNew[Obj, Input, Message] {
// 	tc.msgsSent += len(msgs)
// 	tc.msgs = append(tc.msgs, tc.setMessageSender(sender, msgs)...)
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithMessage(sender NodeID, msg Message) *TestContextNew[Obj, Input, Message] {
// 	tc.msgsSent++
// 	tc.msgs = append(tc.msgs, tc.setMessageSender(sender, []Message{msg})...)
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithMessageChannel(msgCh <-chan Senderany[Message]) *TestContextNew[Obj, Input, Message] {
// 	tc.msgCh = msgCh
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithOutputHandler(outputHandler func(nodeID NodeID, output Output)) *TestContextNew[Obj, Input, Message] {
// 	tc.outputHandler = outputHandler
// 	return tc
// }

// func (tc *TestContextNew[Obj, Input, Message]) WithCall(sender NodeID, call func() []Message) *TestContextNew[Obj, Input, Message] {
// 	msgs := call()
// 	return tc.WithMessages(sender, msgs)
// }

// func (tc *TestContextNew[Obj, Input, Message]) RunUntil(predicate func() bool) {
// 	loop := make(chan bool, 1)
// 	loop <- true
// 	keepLooping := func() {
// 		if len(loop) == 0 {
// 			loop <- true
// 		}
// 	}
// 	for {
// 		select {
// 		case inputs, ok := <-tc.inputCh:
// 			keepLooping()
// 			if !ok {
// 				tc.inputCh = nil
// 				continue
// 			}
// 			if len(inputs) == 0 {
// 				continue
// 			}
// 			for nid, input := range inputs {
// 				tc.inputs[nid] = append(tc.inputs[nid], input)
// 			}
// 			tc.inputCount += len(inputs)
// 		case msg, ok := <-tc.msgCh:
// 			keepLooping()
// 			if !ok {
// 				tc.msgCh = nil
// 				continue
// 			}
// 			tc.msgs = append(tc.msgs, msg)
// 		case <-loop:
// 			if predicate() {
// 				return
// 			}
// 			tc.tryProcessInput()   // Try provide an input, if any and we are lucky.
// 			tc.tryProcessMessage() // Otherwise just process the messages.
// 			if len(tc.msgs) > 0 || tc.inputCount > 0 {
// 				// We can proceed with looping.
// 				loop <- true
// 				continue
// 			}
// 			if tc.inputCh == nil && tc.msgCh == nil {
// 				// Channels are closed and there is no more inputs or messages. Stop it.
// 				return
// 			}
// 			// Otherwise we have to wait for something from channels.
// 		}
// 	}
// }

// func (tc *TestContextNew[Obj, Input, Message]) tryProcessInput() {
// 	if tc.inputCount > 0 && (rand.Float64() <= tc.inputProb || len(tc.msgs) == 0) {
// 		rnd := rand.Intn(tc.inputCount)
// 		var rndNID NodeID
// 		var rndInp Input
// 		for nodeID, nodeInputs := range tc.inputs {
// 			if rnd >= len(nodeInputs) {
// 				rnd -= len(nodeInputs)
// 				continue
// 			}
// 			rndNID = nodeID
// 			rndInp = nodeInputs[0]
// 			tc.inputs[nodeID] = nodeInputs[1:] // Take them in order.
// 			break
// 		}
// 		tc.inputCount--

// 		outMsgs := tc.functors.ApplyInput(tc.nodes[rndNID], rndInp)
// 		newMsgs := tc.setMessageSender(rndNID, outMsgs)
// 		if newMsgs != nil {
// 			tc.msgsSent += len(newMsgs)
// 			tc.msgs = append(tc.msgs, newMsgs...)
// 		}
// 		tc.tryCallOutputHandler(rndNID)
// 	}
// }

// func (tc *TestContextNew[Obj, Input, Message]) tryProcessMessage() {
// 	if len(tc.msgs) == 0 {
// 		return
// 	}
// 	msgIdx := rand.Intn(len(tc.msgs))
// 	msg := tc.msgs[msgIdx]
// 	nid := (*msg.Message).Recipient()
// 	tc.msgs = append(tc.msgs[:msgIdx], tc.msgs[msgIdx+1:]...)
// 	tc.msgsRecv++
// 	if rand.Float64() <= tc.msgDeliveryProb { // Deliver some messages.
// 		gpaMsg := msg.Message
// 		if tc.msgSerialize {
// 			msgBytes := lo.Must(tc.functors.MarshalMessage(*msg.Message))
// 			if m, err := tc.functors.UnmarshalMessage(tc.nodes[nid], msgBytes); err == nil {
// 				gpaMsg = &m
// 				(*gpaMsg).SetSender(msg.Sender)
// 			} else {
// 				// E.g. silent node cannot decode messages.
// 				gpaMsg = nil
// 			}
// 		}
// 		if gpaMsg != nil {
// 			outMsgs := tc.functors.ApplyMessage(tc.nodes[nid], *gpaMsg)
// 			newMsgs := tc.setMessageSender(nid, outMsgs)
// 			if newMsgs != nil {
// 				tc.msgsSent += len(newMsgs)
// 				tc.msgs = append(tc.msgs, newMsgs...)
// 			}
// 			tc.tryCallOutputHandler(nid)
// 		}
// 	}
// }

// func (tc *TestContextNew[Obj, Input, Message]) tryCallOutputHandler(nid NodeID) {
// 	out := tc.functors.Output(tc.nodes[nid])
// 	if out != nil && tc.outputHandler != nil {
// 		tc.outputHandler(nid, out)
// 	}
// }

// func (tc *TestContextNew[Obj, Input, Message]) RunAll() {
// 	tc.RunUntil(tc.OutOfMessagesPredicate())
// }

// // NumberOfOutputs returns a number of non-nil outputs.
// func (tc *TestContextNew[Obj, Input, Message]) NumberOfOutputs() int {
// 	outNum := 0
// 	for _, node := range tc.nodes {
// 		output := tc.functors.Output(node)
// 		if output != nil {
// 			outNum++
// 		}
// 	}
// 	return outNum
// }

// // NumberOfOutputsPredicate runs until there will be at least outNum of non-nil outputs generated.
// func (tc *TestContextNew[Obj, Input, Message]) NumberOfOutputsPredicate(outNum int) func() bool {
// 	return func() bool {
// 		return tc.NumberOfOutputs() >= outNum
// 	}
// }

// // OutOfMessagesPredicate runs until all the messages will be processed.
// func (tc *TestContextNew[Obj, Input, Message]) OutOfMessagesPredicate() func() bool {
// 	return func() bool { return false }
// }

// func (tc *TestContextNew[Obj, Input, Message]) setMessageSender(sender NodeID, msgs []Message) []Senderany[Message] {
// 	if msgs == nil {
// 		return nil
// 	}
// 	result := make([]Senderany[Message], len(msgs))
// 	for i := range msgs {
// 		msgs[i].SetSender(sender)
// 		result[i] = Senderany[Message]{Sender: sender, Message: lo.ToPtr(msgs[i])}
// 	}
// 	return result
// }

// func (tc *TestContextNew[Obj, Input, Message]) PrintAllStatusStrings(prefix string, logFunc func(format string, args ...any)) {
// 	logFunc("TC[%p] Status, |inputs|=%v, inputsCh=%v, |msgs|=%v, msgsCh=%v", tc, tc.inputCount, tc.inputCh != nil, len(tc.msgs), tc.msgCh != nil)
// 	keys := []NodeID{}
// 	for nid := range tc.nodes {
// 		keys = append(keys, nid)
// 	}
// 	// Print them sorted.
// 	sort.Slice(keys, func(i, j int) bool {
// 		return bytes.Compare(keys[i][:], keys[j][:]) < 0
// 	})
// 	for _, nidStr := range keys {
// 		logFunc("TC[%p] %v [node=%v]: %v", tc, prefix, nidStr, tc.functors.StatusString(tc.nodes[nidStr]))
// 	}
// }

// type Senderany[Msg any] struct {
// 	Sender  NodeID
// 	Message *Msg
// }
