package gpa_test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestMsgWrapper(t *testing.T) {
	const wrappedMsgID = 1

	wrapper := gpa.NewMsgWrapper(wrappedMsgID, func(subsystem byte, index int) (gpa.GPA, error) {
		switch subsystem {
		case 2:
			switch index {
			case 3:
				return &subsystemGPA1{}, nil
			}
		case 4:
			switch index {
			case 5:
				return &subsystemGPA2{}, nil
			}
		}

		return nil, fmt.Errorf("unknown subsystem %d index %d", subsystem, index)
	})

	msg1 := &TestWrappedMessage1{V: 42}
	msg2 := &TestWrappedMessage2{V: "hello"}
	wrapped1 := wrapper.WrapMessage(2, 3, msg1)
	wrapped2 := wrapper.WrapMessage(4, 5, msg2)
	unknownSubsystem := wrapper.WrapMessage(2, 4, msg1)
	wrongSubsystem := wrapper.WrapMessage(2, 3, msg2)

	wrapped1Enc := lo.Must(wrapper.MarshalMessage(wrapped1))
	wrapped2Enc := lo.Must(wrapper.MarshalMessage(wrapped2))

	_, err := wrapper.MarshalMessage(unknownSubsystem)
	require.Error(t, err)
	_, err = wrapper.MarshalMessage(wrongSubsystem)
	require.Error(t, err)

	unwrapped1, err := wrapper.UnmarshalMessage(wrapped1Enc)
	require.NoError(t, err)
	require.Equal(t, msg1, unwrapped1.(*gpa.WrappingMsg).Wrapped())

	unwrapped2, err := wrapper.UnmarshalMessage(wrapped2Enc)
	require.NoError(t, err)
	require.Equal(t, msg2, unwrapped2.(*gpa.WrappingMsg).Wrapped())
}

type subsystemGPA1 struct {
	testGPABase[*TestWrappedMessage1]
}

func (g *subsystemGPA1) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return bcs.Unmarshal[*TestWrappedMessage1](data)
}

type TestWrappedMessage1 struct {
	gpa.BasicMessage
	V int
}

type subsystemGPA2 struct {
	testGPABase[*TestWrappedMessage2]
}

func (g *subsystemGPA2) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return bcs.Unmarshal[*TestWrappedMessage2](data)
}

type TestWrappedMessage2 struct {
	gpa.BasicMessage
	V string
}

type testGPABase[MsgType gpa.Message] struct {
}

func (testGPABase[_]) Input(inp gpa.Input) gpa.OutMessages     { return nil }
func (testGPABase[_]) Message(msg gpa.Message) gpa.OutMessages { return nil }
func (testGPABase[_]) Output() gpa.Output                      { return nil }
func (testGPABase[_]) StatusString() string                    { return "" }

func (testGPABase[MsgType]) MarshalMessage(msg gpa.Message) ([]byte, error) {
	m, ok := msg.(MsgType)
	if !ok {
		return nil, fmt.Errorf("unexpected message type %T", msg)
	}

	return bcs.Marshal(&m)
}

func (testGPABase[MsgType]) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return bcs.Unmarshal[MsgType](data)
}
