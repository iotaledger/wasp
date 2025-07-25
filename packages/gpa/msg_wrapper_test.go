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

	msg1 := &TestWrappedMessage1{V: 42}
	msg2 := &TestWrappedMessage2{V: "hello"}
	wrapped1 := wrapper.WrapMessage(2, 3, msg1)
	wrapped2 := wrapper.WrapMessage(4, 5, msg2)

	wrapped1Enc := bcs.MustMarshal(lo.ToPtr[any](wrapped1))
	wrapped2Enc := bcs.MustMarshal(lo.ToPtr[any](wrapped2))

	unwrapped1, err := wrapper.UnmarshalMessage(wrapped1Enc)
	require.NoError(t, err)
	require.Equal(t, msg1, unwrapped1.(*gpa.WrappingMsg).Wrapped())

	unwrapped2, err := wrapper.UnmarshalMessage(wrapped2Enc)
	require.NoError(t, err)
	require.Equal(t, msg2, unwrapped2.(*gpa.WrappingMsg).Wrapped())

	unknownSubsystem := wrapper.WrapMessage(2, 4, msg1)
	wrongSubsystem := wrapper.WrapMessage(2, 3, msg2)

	unknownSubsystemEnc := bcs.MustMarshal(lo.ToPtr[any](unknownSubsystem))
	wrongSubsystemEnc := bcs.MustMarshal(lo.ToPtr[any](wrongSubsystem))

	_, err = wrapper.UnmarshalMessage(unknownSubsystemEnc)
	require.Error(t, err)
	_, err = wrapper.UnmarshalMessage(wrongSubsystemEnc)
	require.Error(t, err)
}

type subsystemGPA1 struct {
	testGPABase[*TestWrappedMessage1]
}

func (g *subsystemGPA1) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data,
		gpa.Mapper{
			1: func() gpa.Message { return new(TestWrappedMessage1) },
		},
	)
}

type TestWrappedMessage1 struct {
	gpa.BasicMessage
	V int
}

func (m *TestWrappedMessage1) MsgType() gpa.MessageType {
	return 1
}

type subsystemGPA2 struct {
	testGPABase[*TestWrappedMessage2]
}

func (g *subsystemGPA2) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data,
		gpa.Mapper{
			2: func() gpa.Message { return new(TestWrappedMessage2) },
		},
	)
}

type TestWrappedMessage2 struct {
	gpa.BasicMessage
	V string
}

func (m *TestWrappedMessage2) MsgType() gpa.MessageType {
	return 2
}

type testGPABase[MsgType gpa.Message] struct{}

func (testGPABase[_]) Input(inp gpa.Input) gpa.OutMessages     { return nil }
func (testGPABase[_]) Message(msg gpa.Message) gpa.OutMessages { return nil }
func (testGPABase[_]) Output() gpa.Output                      { return nil }
func (testGPABase[_]) StatusString() string                    { return "" }
