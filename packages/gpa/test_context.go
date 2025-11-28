// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"

	"github.com/samber/lo"
)

type pendingMessage struct {
	Recipient NodeID
	Msg       MessageIn[any]
}

// TestContext imitates a cluster of nodes and the medium performing the message exchange.
// Inputs are processes in-order for each node individually.
type TestContext[Obj any] struct {
	functors        TestContextFunctors[Obj]
	nodes           map[NodeID]Obj                     // Nodes to test.
	inputs          map[NodeID][]Input                 // Not yet provided inputs.
	inputCh         <-chan map[NodeID]Input            // A way to provide additional inputs w/o synchronizing other parts.
	inputProb       float64                            // A probability to process input, instead of a message (if any).
	inputCount      int                                // Number if inputs still not delivered.
	outputHandler   func(nodeID NodeID, output Output) // User can check outputs w/o synchronizing other parts.
	msgDeliveryProb float64                            // A probability to deliver a message (to not discard/loose it).
	msgSerialize    bool                               // Use serialization/deserialization when delivering the messages?
	msgs            []pendingMessage                   // Not yet delivered messages.
	msgsSent        int                                // Stats.
	msgsRecv        int                                // Stats.
	bytesRecv       int
}

func NewTestContext[Obj any](nodes map[NodeID]Obj, functors ...TestContextFunctors[Obj]) *TestContext[Obj] {
	inputs := map[NodeID][]Input{}
	for n := range nodes {
		inputs[n] = []Input{}
	}

	if len(functors) == 0 {
		functors = append(functors, TestContextFunctors[Obj]{})
	}
	setDefaultFunctors(&functors[0])

	tc := TestContext[Obj]{
		functors:        functors[0],
		msgSerialize:    true,
		nodes:           nodes,
		inputs:          inputs,
		inputProb:       1.0,
		inputCount:      0,
		msgDeliveryProb: 1.0,
		msgs:            []pendingMessage{},
	}
	return &tc
}

func (tc *TestContext[Obj]) WithoutSerialization() *TestContext[Obj] {
	tc.msgSerialize = false
	return tc
}

func (tc *TestContext[Obj]) MsgCounts() (int, int) {
	return tc.msgsSent, tc.msgsRecv
}

// AddInputs adds new inputs to the existing set.
// The inputs will be overridden, if exist for the same nodes.
func (tc *TestContext[Obj]) AddInputs(inputs map[NodeID]Input) {
	for nid := range inputs {
		tc.inputs[nid] = append(tc.inputs[nid], inputs[nid])
	}
	tc.inputCount += len(inputs)
}

func (tc *TestContext[Obj]) WithInput(nodeID NodeID, input Input) *TestContext[Obj] {
	tc.AddInputs(map[NodeID]Input{nodeID: input})
	return tc
}

func (tc *TestContext[Obj]) WithInputs(inputs map[NodeID]Input) *TestContext[Obj] {
	tc.AddInputs(inputs)
	return tc
}

func (tc *TestContext[Obj]) WithInputChannel(inputCh <-chan map[NodeID]Input) *TestContext[Obj] {
	tc.inputCh = inputCh
	return tc
}

func (tc *TestContext[Obj]) WithInputProbability(inputProb float64) *TestContext[Obj] {
	tc.inputProb = inputProb
	return tc
}

func (tc *TestContext[Obj]) WithMessageDeliveryProbability(msgDeliveryProb float64) *TestContext[Obj] {
	tc.msgDeliveryProb = msgDeliveryProb
	return tc
}

func (tc *TestContext[Obj]) WithMessages(recipient NodeID, msgs []MessageIn[any]) *TestContext[Obj] {
	tc.addMessages(lo.Map(msgs, func(m MessageIn[any], _ int) pendingMessage {
		return pendingMessage{Recipient: recipient, Msg: m}
	}))
	return tc
}

func (tc *TestContext[Obj]) addMessages(msgs []pendingMessage) {
	tc.msgsSent += len(msgs)
	tc.msgs = append(tc.msgs, msgs...)
}

func (tc *TestContext[Obj]) WithMessage(recipient NodeID, msg MessageIn[any]) *TestContext[Obj] {
	tc.msgsSent++
	tc.msgs = append(tc.msgs, pendingMessage{Recipient: recipient, Msg: msg})
	return tc
}

func (tc *TestContext[Obj]) WithOutputHandler(outputHandler func(nodeID NodeID, output Output)) *TestContext[Obj] {
	tc.outputHandler = outputHandler
	return tc
}

func (tc *TestContext[Obj]) RunUntil(predicate func() bool) {
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

func (tc *TestContext[Obj]) tryProcessInput() {
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
		tc.addMessages(lo.Map(msgs, func(m MessageOut, _ int) pendingMessage {
			return pendingMessage{Recipient: m.Recipient, Msg: NewMessageIn(rndNID, m.Payload)}
		}))
		tc.tryCallOutputHandler(rndNID)
	}
}

func (tc *TestContext[Obj]) tryProcessMessage() {
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
		msgBytes := lo.Must(tc.functors.MarshalPayload(tc.nodes[msg.Sender], msg.Payload))
		tc.bytesRecv += len(msgBytes)
		m, err := tc.functors.UnmarshalPayload(tc.nodes[nid], msgBytes)
		if err != nil {
			// E.g. silent node cannot decode messages.
			return
		}
		msg = NewMessageIn(msg.Sender, m)
	}
	// fmt.Printf("%s -> %s :: %s (count: %d / %d bytes)\n", msg.Sender.ShortString(), nid.ShortString(), msg.Payload, tc.msgsRecv, tc.bytesRecv)
	msgs := tc.functors.ApplyMessage(tc.nodes[nid], msg)
	tc.addMessages(lo.Map(msgs, func(m TypedMessageOut[any], _ int) pendingMessage {
		return pendingMessage{Recipient: m.Recipient, Msg: NewMessageIn(nid, m.Payload)}
	}))
	tc.tryCallOutputHandler(nid)
}

func (tc *TestContext[Obj]) tryCallOutputHandler(nid NodeID) {
	out := tc.functors.Output(tc.nodes[nid])
	if out != nil && tc.outputHandler != nil {
		tc.outputHandler(nid, out)
	}
}

func (tc *TestContext[Obj]) RunAll() {
	tc.RunUntil(tc.OutOfMessagesPredicate())
}

// NumberOfOutputs returns a number of non-nil outputs.
func (tc *TestContext[Obj]) NumberOfOutputs() int {
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
func (tc *TestContext[Obj]) NumberOfOutputsPredicate(outNum int) func() bool {
	return func() bool {
		return tc.NumberOfOutputs() >= outNum
	}
}

// OutOfMessagesPredicate runs until all the messages will be processed.
func (tc *TestContext[Obj]) OutOfMessagesPredicate() func() bool {
	return func() bool { return false }
}

func (tc *TestContext[Obj]) PrintAllStatusStrings(prefix string, logFunc func(format string, args ...any)) {
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

func ToAnyPayloadsOut[Payload any](payloads []TypedMessageOut[Payload]) []MessageOut {
	res := make([]MessageOut, len(payloads))
	for i, p := range payloads {
		res[i] = MessageOut{
			Recipient: p.Recipient,
			Payload:   p.Payload,
		}
	}
	return res
}

func FindAndInvokeInputHandler(obj any, input Input) []MessageOut {
	type genericInputHandler interface {
		Input(input Input) []MessageOut
	}
	if handler, ok := obj.(genericInputHandler); ok {
		return handler.Input(input)
	}

	objV := reflect.ValueOf(obj)
	objT := objV.Type()

	handlerMethodT := reflect.FuncOf([]reflect.Type{objT, reflect.TypeOf(input)}, []reflect.Type{reflect.TypeOf([]MessageOut{})}, false)
	handlerPtrMethodT := reflect.FuncOf([]reflect.Type{objT, reflect.PtrTo(reflect.TypeOf(input))}, []reflect.Type{reflect.TypeOf([]MessageOut{})}, false)

	for i := 0; i < objT.NumMethod(); i++ {
		methodT := objT.Method(i)
		if methodT.Type == handlerMethodT {
			result := methodT.Func.Call([]reflect.Value{objV, reflect.ValueOf(input)})
			return result[0].Interface().([]MessageOut)
		}
		if methodT.Type == handlerPtrMethodT {
			// TODO: copy non-addressable value
			result := methodT.Func.Call([]reflect.Value{objV, reflect.ValueOf(input).Addr()})
			return result[0].Interface().([]MessageOut)
		}
	}

	panic(fmt.Errorf("no input handler found with signature %v or %v", handlerMethodT, handlerPtrMethodT))
}

func FindAndInvokeMessageHandler(obj any, msg MessageIn[any]) []MessageOut {
	type genericMessageHandler interface {
		Message(msg MessageIn[any]) []MessageOut
	}
	if handler, ok := obj.(genericMessageHandler); ok {
		return handler.Message(msg)
	}

	// There is no good instruments with work with generics in Go reflection.
	// So we are forced to fallback to name-based heuristics.
	// We could just hard-code name, but then tests would break after renamings. So we dynamically get current name of type.
	samplePayloadWithKey := PayloadWithKey[struct{}, struct{}]{}

	switch {
	case isSameGenericType(reflect.TypeOf(msg.Payload), reflect.TypeOf(samplePayloadWithKey)):
		return findAndInvokePayloadWithKeyMessageHandler(obj, msg)
	default:
		return findAndInvokeSimpleMessageHandler(obj, msg)
	}
}

func findAndInvokeSimpleMessageHandler(obj any, msg MessageIn[any]) []MessageOut {
	objV := reflect.ValueOf(obj)
	objT := objV.Type()
	msgV := reflect.ValueOf(msg)

	for i := 0; i < objT.NumMethod(); i++ {
		methodT := objT.Method(i)

		if methodT.Type.NumIn() != 2 || methodT.Type.NumOut() != 1 {
			continue
		}
		if methodT.Type.Out(0) != reflect.TypeOf([]MessageOut{}) {
			continue
		}
		if methodT.Type.In(0) != objT {
			panic(fmt.Errorf("mismatched receiver type: %v != %v", methodT.Type.In(0), objT))
		}

		msgArgT := methodT.Type.In(1)
		isMsgIn, payloadArgT := isMessageInType(msgArgT)
		if !isMsgIn || payloadArgT.Type != reflect.TypeOf(msg.Payload) {
			continue
		}

		// We have a match.
		convertedMsgV := reflect.New(msgArgT).Elem()

		for i := 0; i < msgArgT.NumField(); i++ {
			fieldV := msgV.Field(i)
			destFieldV := convertedMsgV.Field(i)

			if i == payloadArgT.Index[0] {
				convertedFieldV := fieldV.Elem().Convert(destFieldV.Type())
				destFieldV.Set(convertedFieldV)
			} else {
				destFieldV.Set(fieldV)
			}
		}

		result := methodT.Func.Call([]reflect.Value{objV, convertedMsgV})
		return result[0].Interface().([]MessageOut)
	}

	panic(fmt.Errorf("no message handler found for message with payload type %T in object of type %T", msg.Payload, obj))
}

func findAndInvokePayloadWithKeyMessageHandler(obj any, msg MessageIn[any]) []MessageOut {
	objV := reflect.ValueOf(obj)
	objT := objV.Type()
	msgV := reflect.ValueOf(msg)
	payloadWithKeyV := reflect.ValueOf(msg.Payload)
	payloadWithKeyT := payloadWithKeyV.Type()
	keyV := payloadWithKeyV.Field(getPayloadWithKeyKeyFieldIndex(payloadWithKeyT))
	payloadV := payloadWithKeyV.Field(getPayloadWithKeyPayloadFieldIndex(payloadWithKeyT)).Elem()
	payloadT := payloadV.Type()

	for i := 0; i < objT.NumMethod(); i++ {
		methodT := objT.Method(i)

		if methodT.Type.NumIn() != 3 || methodT.Type.NumOut() != 1 {
			continue
		}
		if methodT.Type.Out(0) != reflect.TypeOf([]MessageOut{}) {
			continue
		}
		if methodT.Type.In(0) != objT {
			panic(fmt.Errorf("mismatched receiver type: %v != %v", methodT.Type.In(0), objT))
		}

		keyArgT := methodT.Type.In(1)
		if keyArgT != keyV.Type() {
			fmt.Println("key type mismatch:", methodT.Name, keyArgT, keyV.Type())
			continue
		}

		msgArgT := methodT.Type.In(2)
		isMsgIn, payloadArgT := isMessageInType(msgArgT)
		if !isMsgIn || payloadArgT.Type != payloadT {
			fmt.Println("payload type mismatch:", methodT.Name, payloadArgT.Type, payloadT)
			continue
		}

		// We have a match.
		convertedMsgV := reflect.New(msgArgT).Elem()

		for i := 0; i < msgArgT.NumField(); i++ {
			fieldV := msgV.Field(i)
			destFieldV := convertedMsgV.Field(i)

			if i == payloadArgT.Index[0] {
				convertedFieldV := payloadV.Convert(destFieldV.Type())
				destFieldV.Set(convertedFieldV)
			} else {
				destFieldV.Set(fieldV)
			}
		}

		result := methodT.Func.Call([]reflect.Value{objV, keyV, convertedMsgV})
		return result[0].Interface().([]MessageOut)
	}

	panic(fmt.Errorf("no message handler found for message with payload type %T in object of type %T", msg.Payload, obj))
}

func isMessageInType(t reflect.Type) (isMessageIn bool, payloadField reflect.StructField) {
	// There is no good instruments with work with generics in Go reflection.
	// So we are forced to fallback to name-based heuristics.
	// We could just hard-code name, but then tests would break after renamings. So we dynamically get current name of type.
	type privatePayloadType struct{}
	sampleTypeT := reflect.TypeOf(MessageIn[privatePayloadType]{})

	if !isSameGenericType(t, sampleTypeT) {
		return false, reflect.StructField{}
	}

	// We also dynamically find Payload field - also to handle future renamings.
	payloadFieldIndex := getFieldIndexOfType(sampleTypeT, reflect.TypeOf(privatePayloadType{}))

	return true, t.Field(payloadFieldIndex)
}

func isSameGenericType(t1, t2 reflect.Type) bool {
	name1 := getGenericTypeName(t1)
	if name1 == "" {
		return false // not a generic type
	}
	name2 := getGenericTypeName(t2)
	if name2 == "" {
		return false // not a generic type
	}
	return name1 == name2
}

func getGenericTypeName(t reflect.Type) string {
	paramsStart := strings.Index(t.Name(), "[")
	if paramsStart < 0 {
		// not a generic type
		return ""
	}
	return t.Name()[:paramsStart]
}

func getFieldIndexOfType(structType reflect.Type, fieldType reflect.Type) int {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Type == fieldType {
			return i
		}
	}
	panic(fmt.Errorf("no field of type %v found in struct %v", fieldType, structType))
}

// func getPayloadWithKeySubsystemIDFieldIndex(structType reflect.Type) int {
// 	samplePayloadWithKey := PayloadWithKey[struct{}, struct{}]{}
// 	const fieldName = "SubsystemID"
// 	f, found := reflect.TypeOf(samplePayloadWithKey).FieldByName(fieldName)
// 	if !found {
// 		panic(fmt.Errorf("no %v field found in %T", fieldName, samplePayloadWithKey))
// 	}
// 	return f.Index[0]
// }

func getPayloadWithKeyKeyFieldIndex(structType reflect.Type) int {
	type privateKeyType struct{}
	samplePayloadWithKey := PayloadWithKey[privateKeyType, struct{}]{}
	return getFieldIndexOfType(reflect.TypeOf(samplePayloadWithKey), reflect.TypeOf(privateKeyType{}))
}

func getPayloadWithKeyPayloadFieldIndex(structType reflect.Type) int {
	type privatePayloadType struct{}
	samplePayloadWithKey := PayloadWithKey[struct{}, privatePayloadType]{}
	return getFieldIndexOfType(reflect.TypeOf(samplePayloadWithKey), reflect.TypeOf(privatePayloadType{}))
}

type TestContextFunctors[Obj any] struct {
	ApplyInput       func(obj Obj, input Input) []MessageOut
	ApplyMessage     func(obj Obj, msg MessageIn[any]) []MessageOut
	Output           func(obj Obj) any
	StatusString     func(obj Obj) string
	MarshalPayload   func(obj Obj, msg any) ([]byte, error)
	UnmarshalPayload func(obj Obj, data []byte) (any, error)
}

func setDefaultFunctors[Obj any](functors *TestContextFunctors[Obj]) {
	if functors.ApplyInput == nil {
		functors.ApplyInput = func(obj Obj, input Input) []MessageOut {
			return FindAndInvokeInputHandler(obj, input)
		}
	}
	if functors.ApplyMessage == nil {
		functors.ApplyMessage = func(obj Obj, msg MessageIn[any]) []MessageOut {
			return FindAndInvokeMessageHandler(obj, msg)
		}
	}
	if functors.MarshalPayload == nil {
		type marshaler interface {
			MarshalPayload(msg any) ([]byte, error)
		}
		functors.MarshalPayload = func(obj Obj, msg any) ([]byte, error) {
			return interface{}(obj).(marshaler).MarshalPayload(msg)
		}
	}
	if functors.UnmarshalPayload == nil {
		type unmarshaler interface {
			UnmarshalPayload(data []byte) (any, error)
		}
		functors.UnmarshalPayload = func(obj Obj, data []byte) (any, error) {
			return interface{}(obj).(unmarshaler).UnmarshalPayload(data)
		}
	}
	if functors.Output == nil {
		functors.Output = func(obj Obj) any {
			type outputter interface {
				Output() Output
			}
			return interface{}(obj).(outputter).Output()
		}
	}
	if functors.StatusString == nil {
		functors.StatusString = func(obj Obj) string {
			type statusStringer interface {
				StatusString() string
			}
			return interface{}(obj).(statusStringer).StatusString()
		}
	}
}
