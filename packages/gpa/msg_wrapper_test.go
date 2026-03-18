package gpa_test

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

func TestMsgWrapper(t *testing.T) {
	const wrappedMsgID = 1

	wrapper := gpa.NewMsgWrapper(wrappedMsgID, func(subsystem byte, index int) (gpa.GPA, error) {
		switch subsystem {
		case 2:
			if index == 3 {
				return &subsystemGPA1{}, nil
			}
		case 4:
			if index == 5 {
				return &subsystemGPA2{}, nil
			}
		}

		return nil, fmt.Errorf("unknown subsystem %d index %d", subsystem, index)
	})

	sender := gpa.NodeID{1}
	recipient := gpa.NodeID{2}

	msg1 := gpa.NewMessageIn(sender, &TestWrappedMessage1{V: 42})
	msg2 := gpa.NewMessageOut(recipient, &TestWrappedMessage2{V: "hello"})

	wrapped1 := wrapper.WrapMessageIn(2, 3, msg1)
	wrapped2 := wrapper.WrapMessageOut(4, 5, msg2)

	wrapped1Enc := bcs.MustMarshal(lo.ToPtr[any](wrapped1.Payload))
	wrapped2Enc := bcs.MustMarshal(lo.ToPtr[any](wrapped2.Payload))

	unwrapped1, err := wrapper.UnmarshalPayload(wrapped1Enc)
	require.NoError(t, err)
	require.Equal(t, msg1, unwrapped1.(*gpa.WrappingMsg).WrappedIn(msg1.Sender))

	unwrapped2, err := wrapper.UnmarshalPayload(wrapped2Enc)
	require.NoError(t, err)
	require.Equal(t, msg2, unwrapped2.(*gpa.WrappingMsg).WrappedOut(msg2.Recipient))

	unknownSubsystem := wrapper.WrapMessageIn(2, 4, msg1)
	wrongSubsystem := wrapper.WrapMessageOut(2, 3, msg2)

	unknownSubsystemEnc := bcs.MustMarshal(lo.ToPtr[any](unknownSubsystem.Payload))
	wrongSubsystemEnc := bcs.MustMarshal(lo.ToPtr[any](wrongSubsystem.Payload))

	_, err = wrapper.UnmarshalPayload(unknownSubsystemEnc)
	require.Error(t, err)
	_, err = wrapper.UnmarshalPayload(wrongSubsystemEnc)
	require.Error(t, err)
}

type subsystemGPA1 struct {
	testGPABase[*TestWrappedMessage1]
}

func (g *subsystemGPA1) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data,
		gpa.PayloadAllocator{
			1: func() gpa.MessagePayload { return &TestWrappedMessage1{} },
		},
	)
}

type TestWrappedMessage1 struct {
	V int
}

func (m *TestWrappedMessage1) MsgType() gpa.MessageType {
	return 1
}

type subsystemGPA2 struct {
	testGPABase[*TestWrappedMessage2]
}

func (g *subsystemGPA2) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data,
		gpa.PayloadAllocator{
			2: func() gpa.MessagePayload { return &TestWrappedMessage2{} },
		},
	)
}

type TestWrappedMessage2 struct {
	V string
}

func (m *TestWrappedMessage2) MsgType() gpa.MessageType {
	return 2
}

type testGPABase[MsgType gpa.MessagePayload] struct{}

func (testGPABase[_]) Input(inp gpa.Input) []gpa.MessageOut       { return nil }
func (testGPABase[_]) Message(msg gpa.MessageIn) []gpa.MessageOut { return nil }
func (testGPABase[_]) Output() gpa.Output                         { return nil }
func (testGPABase[_]) StatusString() string                       { return "" }
